package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	quickjs "github.com/buke/quickjs-go"
	"github.com/patrickmoriarty/api-cli/internal/request"
	"github.com/patrickmoriarty/api-cli/internal/secret"
)

type ExecOptions struct {
	Provider string
	Profile  string
	Command  string
	Params   map[string]string
	Env      map[string]string
	Timeout  time.Duration
}

type ExecResult struct {
	Status int
	Body   string
	JSON   string
}

type CommandDoc struct {
	Name string
	Desc string
	Args []string
}

func Execute(script []byte, opts ExecOptions) (*ExecResult, error) {
	if len(script) == 0 {
		return nil, errors.New("provider script is empty")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 20 * time.Second
	}

	execTimeout := uint64(opts.Timeout.Seconds())
	if execTimeout == 0 {
		execTimeout = 1
	}

	rt := quickjs.NewRuntime(
		quickjs.WithExecuteTimeout(execTimeout),
		quickjs.WithMemoryLimit(128*1024*1024),
	)
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	ctx.Globals().Set("fetch", ctx.NewFunction(fetchFunc(opts.Timeout)))
	ctx.Globals().Set("secret", ctx.NewFunction(secretFunc(opts.Provider, opts.Profile)))
	ctx.Globals().Set("env", ctx.NewFunction(envFunc(opts.Env)))
	ctx.Globals().Set("sleep", ctx.NewFunction(sleepFunc()))
	ctx.Globals().Set("provider", ctx.NewString(opts.Provider))
	ctx.Globals().Set("profile", ctx.NewString(opts.Profile))
	ctx.Globals().Set("params", mapToObject(ctx, opts.Params))

	source := prepareScript(string(script))
	val := ctx.Eval(source)
	defer val.Free()
	if val.IsException() {
		return nil, ctx.Exception()
	}
	defaultVal := ctx.Globals().Get("__api_default__")
	defer defaultVal.Free()
	if defaultVal.IsUndefined() || defaultVal.IsNull() {
		return nil, errors.New("script did not set export default")
	}
	if !defaultVal.IsObject() {
		return nil, errors.New("default export must be an object")
	}

	resultVal, err := invokeCommand(ctx, defaultVal, opts.Command, opts.Params)
	if err != nil {
		return nil, err
	}
	defer resultVal.Free()

	res := &ExecResult{}
	if resultVal.IsObject() {
		statusVal := resultVal.Get("status")
		bodyVal := resultVal.Get("body")
		jsonVal := resultVal.Get("json")
		defer statusVal.Free()
		defer bodyVal.Free()
		defer jsonVal.Free()
		if !statusVal.IsUndefined() {
			res.Status = int(statusVal.ToInt32())
		}
		if !bodyVal.IsUndefined() && !bodyVal.IsNull() {
			res.Body = bodyVal.ToString()
		}
		if !jsonVal.IsUndefined() && !jsonVal.IsNull() {
			res.JSON = jsonVal.JSONStringify()
		}
	}

	if res.JSON == "" && resultVal.IsObject() {
		res.JSON = resultVal.JSONStringify()
	}

	if res.Body == "" && !resultVal.IsUndefined() {
		res.Body = resultVal.ToString()
	}

	return res, nil
}

func ListCommands(script []byte) ([]string, error) {
	docs, err := DescribeCommands(script)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.Name)
	}
	return out, nil
}

func DescribeCommands(script []byte) ([]CommandDoc, error) {
	rt := quickjs.NewRuntime(
		quickjs.WithExecuteTimeout(2),
		quickjs.WithMemoryLimit(64*1024*1024),
	)
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	source := prepareScript(string(script))
	val := ctx.Eval(source)
	defer val.Free()
	if val.IsException() {
		return nil, ctx.Exception()
	}
	defaultVal := ctx.Globals().Get("__api_default__")
	defer defaultVal.Free()
	if defaultVal.IsUndefined() || defaultVal.IsNull() {
		return nil, errors.New("script did not set export default")
	}
	if !defaultVal.IsObject() {
		return nil, errors.New("default export must be an object")
	}
	keysFn := ctx.Eval("(obj) => Object.keys(obj || {})")
	defer keysFn.Free()
	if keysFn.IsException() {
		return nil, ctx.Exception()
	}
	keys := keysFn.Execute(ctx.NewUndefined(), defaultVal)
	defer keys.Free()
	if keys.IsException() {
		return nil, ctx.Exception()
	}
	var out []CommandDoc
	if keys.IsArray() {
		length := keys.Get("length")
		defer length.Free()
		for i := int64(0); i < length.ToInt64(); i++ {
			key := keys.GetIdx(i)
			name := key.ToString()
			key.Free()

			entry := defaultVal.Get(name)
			if entry == nil || entry.IsUndefined() || entry.IsNull() {
				continue
			}
			descVal := entry.Get("desc")
			argsVal := entry.Get("args")
			doc := CommandDoc{Name: name}
			if descVal != nil && !descVal.IsUndefined() && !descVal.IsNull() {
				doc.Desc = descVal.ToString()
			}
			if argsVal != nil && argsVal.IsArray() {
				length := argsVal.Get("length")
				for j := int64(0); j < length.ToInt64(); j++ {
					item := argsVal.GetIdx(j)
					doc.Args = append(doc.Args, item.ToString())
					item.Free()
				}
				length.Free()
			}
			if descVal != nil {
				descVal.Free()
			}
			if argsVal != nil {
				argsVal.Free()
			}
			entry.Free()
			out = append(out, doc)
		}
	}
	return out, nil
}

func prepareScript(source string) string {
	if strings.Contains(source, "export default") {
		return strings.Replace(source, "export default", "globalThis.__api_default__ =", 1)
	}
	return source
}

func invokeCommand(ctx *quickjs.Context, defaultVal *quickjs.Value, command string, params map[string]string) (*quickjs.Value, error) {
	entry := defaultVal.Get(command)
	defer entry.Free()
	if entry.IsUndefined() || entry.IsNull() {
		return nil, fmt.Errorf("command not found: %s", command)
	}
	fn := entry.Get("run")
	defer fn.Free()
	if fn.IsUndefined() || fn.IsNull() || !fn.IsFunction() {
		return nil, fmt.Errorf("command missing run(): %s", command)
	}
	paramsVal := mapToObject(ctx, params)
	defer paramsVal.Free()
	result := fn.Execute(ctx.NewUndefined(), paramsVal)
	if result.IsException() {
		defer result.Free()
		return nil, ctx.Exception()
	}
	return result, nil
}

func mapToObject(ctx *quickjs.Context, values map[string]string) *quickjs.Value {
	obj := ctx.NewObject()
	for k, v := range values {
		obj.Set(k, ctx.NewString(v))
	}
	return obj
}

func fetchFunc(timeout time.Duration) func(*quickjs.Context, *quickjs.Value, []*quickjs.Value) *quickjs.Value {
	return func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if len(args) == 0 {
			return ctx.ThrowInternalError("fetch expects a url or options object")
		}
		var spec request.Spec
		switch len(args) {
		case 1:
			if args[0].IsString() {
				spec.URL = args[0].ToString()
			} else {
				var err error
				spec, err = specFromValue(args[0])
				if err != nil {
					return ctx.ThrowInternalError("invalid fetch options: %v", err)
				}
			}
		default:
			if !args[0].IsString() {
				return ctx.ThrowInternalError("fetch url must be a string")
			}
			spec.URL = args[0].ToString()
			opts, err := specFromValue(args[1])
			if err != nil {
				return ctx.ThrowInternalError("invalid fetch options: %v", err)
			}
			if opts.Method != "" {
				spec.Method = opts.Method
			}
			if opts.Headers != nil {
				spec.Headers = opts.Headers
			}
			if opts.Body != nil {
				spec.Body = opts.Body
			}
		}
		if spec.URL == "" {
			return ctx.ThrowInternalError("fetch url is required")
		}
		resp, err := request.Do(spec, timeout)
		if err != nil {
			return ctx.ThrowInternalError("request failed: %v", err)
		}
		obj := ctx.NewObject()
		obj.Set("status", ctx.NewInt32(int32(resp.Status)))
		headers := ctx.NewObject()
		for k, v := range resp.Headers {
			if len(v) > 0 {
				headers.Set(k, ctx.NewString(v[0]))
			}
		}
		obj.Set("headers", headers)
		obj.Set("body", ctx.NewString(string(resp.Body)))
		if resp.ParsedJSON != nil {
			b, _ := json.Marshal(resp.ParsedJSON)
			obj.Set("json", ctx.ParseJSON(string(b)))
		} else {
			obj.Set("json", ctx.NewNull())
		}
		return obj
	}
}

func specFromValue(val *quickjs.Value) (request.Spec, error) {
	payload := val.JSONStringify()
	if payload == "" || payload == "undefined" {
		return request.Spec{}, errors.New("fetch options are empty")
	}
	var spec request.Spec
	if err := json.Unmarshal([]byte(payload), &spec); err != nil {
		return request.Spec{}, err
	}
	return spec, nil
}

func secretFunc(provider, profile string) func(*quickjs.Context, *quickjs.Value, []*quickjs.Value) *quickjs.Value {
	return func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if len(args) == 0 {
			return ctx.ThrowInternalError("secret expects a name")
		}
		name := args[0].ToString()
		val, err := secret.Get(provider, profile, name)
		if err != nil {
			return ctx.ThrowInternalError("secret not found: %s", name)
		}
		return ctx.NewString(val)
	}
}

func envFunc(values map[string]string) func(*quickjs.Context, *quickjs.Value, []*quickjs.Value) *quickjs.Value {
	return func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if len(args) == 0 {
			return ctx.ThrowInternalError("env expects a name")
		}
		name := args[0].ToString()
		val := ""
		if values != nil {
			if v, ok := values[name]; ok {
				val = v
			}
		}
		if val == "" {
			val = os.Getenv(name)
		}
		return ctx.NewString(val)
	}
}

func sleepFunc() func(*quickjs.Context, *quickjs.Value, []*quickjs.Value) *quickjs.Value {
	return func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if len(args) == 0 {
			return ctx.ThrowInternalError("sleep expects milliseconds")
		}
		ms := args[0].ToFloat64()
		if ms < 0 {
			ms = 0
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return ctx.NewUndefined()
	}
}
