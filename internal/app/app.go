package app

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/patrickjm/api-cli/internal/config"
	"github.com/patrickjm/api-cli/internal/provider"
	"github.com/patrickjm/api-cli/internal/runtime"
	"github.com/patrickjm/api-cli/internal/secret"
	"github.com/spf13/cobra"
)

var (
	configDir string
	profile   string
	timeout   time.Duration
	jsonOut   bool
	version   = "dev"
)

func Execute() error {
	root := newRootCmd()
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "api is a scriptable API client",
		Args:  cobra.ArbitraryArgs,
		RunE:  runProvider,
	}
	cmd.Version = version
	cmd.SetVersionTemplate("api version {{.Version}}\n")
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		base, err := config.BaseDir(configDir)
		if err != nil {
			return err
		}
		if err := config.EnsureLayout(base); err != nil {
			return err
		}
		if err := provider.EnsureDefaults(base); err != nil {
			return err
		}
		return nil
	}
	cmd.PersistentFlags().StringVarP(&configDir, "config", "c", "", "config directory")
	cmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "profile name")
	cmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 20*time.Second, "request timeout")
	cmd.PersistentFlags().BoolVarP(&jsonOut, "json", "j", false, "emit JSON output")
	cmd.PersistentFlags().StringArrayP("param", "s", nil, "request param key=value")

	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newProvidersCmd())
	cmd.AddCommand(newInspectCmd())
	cmd.AddCommand(newEnvCmd())
	cmd.AddCommand(newProfileCmd())
	cmd.AddCommand(newSecretCmd())
	return cmd
}

func runProvider(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Usage()
	}
	providerName, commandName, paramArgs := parseProviderArgs(args)
	params, err := parseParams(paramArgs)
	if err != nil {
		return err
	}
	paramsFlag, _ := cmd.Flags().GetStringArray("param")
	paramsFromFlags, err := parseParams(paramsFlag)
	if err != nil {
		return err
	}
	for k, v := range paramsFromFlags {
		params[k] = v
	}

	base, err := config.BaseDir(configDir)
	if err != nil {
		return err
	}
	if err := config.EnsureLayout(base); err != nil {
		return err
	}
	providerPath := config.ProviderPath(base, providerName)
	script, err := os.ReadFile(providerPath)
	if err != nil {
		return fmt.Errorf("provider not found: %s", providerName)
	}

	profiles, err := config.LoadProfiles(config.ProviderProfilesPath(base, providerName))
	if err != nil {
		return err
	}
	resolvedProfile, err := config.ResolveProfile(profiles, profile)
	if err != nil {
		return err
	}
	env := profiles.Profiles[resolvedProfile].Env

	result, err := runtime.Execute(script, runtime.ExecOptions{
		Provider: providerName,
		Profile:  resolvedProfile,
		Command:  commandName,
		Params:   params,
		Env:      env,
		Timeout:  timeout,
	})
	if err != nil {
		return err
	}

	if result.Status >= 400 {
		if result.Body != "" {
			_, _ = fmt.Fprintln(os.Stderr, result.Body)
		}
		return fmt.Errorf("request failed with status %d", result.Status)
	}

	if jsonOut {
		if result.JSON != "" {
			fmt.Fprintln(cmd.OutOrStdout(), result.JSON)
			return nil
		}
	}
	if result.Body != "" {
		fmt.Fprintln(cmd.OutOrStdout(), result.Body)
	}
	return nil
}

func newInstallCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "install <source>",
		Short: "install a provider script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			contents, err := provider.ReadSource(source, os.Stdin)
			if err != nil {
				return err
			}
			if name == "" {
				name = inferProviderName(source)
			}
			if name == "" {
				return errors.New("provider name is required")
			}
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderPath(base, name)
			if err := provider.SaveProvider(path, contents); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "installed %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "provider name")
	return cmd
}

func newProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "list installed providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			list, err := provider.ListProviders(config.ProvidersDir(base))
			if err != nil {
				return err
			}
			for _, name := range list {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		},
	}
}

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <provider>",
		Short: "list commands for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			script, err := os.ReadFile(config.ProviderPath(base, name))
			if err != nil {
				return err
			}
			commands, err := runtime.DescribeCommands(script)
			if err != nil {
				return err
			}
			for _, c := range commands {
				if len(c.Args) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", c.Name, c.Desc, strings.Join(c.Args, ","))
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", c.Name, c.Desc)
				}
			}
			return nil
		},
	}
}

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "manage provider profiles",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list <provider>",
		Short: "list profiles for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			profiles, err := config.LoadProfiles(config.ProviderProfilesPath(base, args[0]))
			if err != nil {
				return err
			}
			for name := range profiles.Profiles {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "add <provider> <name>",
		Short: "add a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			if err := config.AddProfile(profiles, args[1]); err != nil {
				return err
			}
			return config.SaveProfiles(path, profiles)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "rm <provider> <name>",
		Short: "remove a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			if err := config.RemoveProfile(profiles, args[1]); err != nil {
				return err
			}
			return config.SaveProfiles(path, profiles)
		},
	})
	return cmd
}

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "manage provider environment values",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <provider> <name> <value>",
		Short: "set an environment value",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			config.UpsertEnv(profiles, resolved, args[1], args[2])
			return config.SaveProfiles(path, profiles)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "unset <provider> <name>",
		Short: "remove an environment value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			config.RemoveEnv(profiles, resolved, args[1])
			return config.SaveProfiles(path, profiles)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list <provider>",
		Short: "list environment values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			prof := profiles.Profiles[resolved]
			for key, value := range prof.Env {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, value)
			}
			return nil
		},
	})
	return cmd
}

func newSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "manage secrets",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <provider> <name> <value>",
		Short: "set a secret",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			if err := secret.Set(args[0], resolved, args[1], args[2]); err != nil {
				return err
			}
			config.UpsertSecret(profiles, resolved, args[1])
			return config.SaveProfiles(path, profiles)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "unset <provider> <name>",
		Short: "remove a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			_ = secret.Delete(args[0], resolved, args[1])
			config.RemoveSecret(profiles, resolved, args[1])
			return config.SaveProfiles(path, profiles)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list <provider>",
		Short: "list secrets for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := config.BaseDir(configDir)
			if err != nil {
				return err
			}
			if err := config.EnsureLayout(base); err != nil {
				return err
			}
			path := config.ProviderProfilesPath(base, args[0])
			profiles, err := config.LoadProfiles(path)
			if err != nil {
				return err
			}
			resolved, err := config.ResolveProfile(profiles, profile)
			if err != nil {
				return err
			}
			prof := profiles.Profiles[resolved]
			for _, key := range prof.Secrets {
				fmt.Fprintln(cmd.OutOrStdout(), key)
			}
			return nil
		},
	})
	return cmd
}

func parseProviderArgs(args []string) (string, string, []string) {
	providerArg := args[0]
	command := "default"
	idx := 1
	if strings.Contains(providerArg, ".") {
		parts := strings.SplitN(providerArg, ".", 2)
		providerArg = parts[0]
		command = parts[1]
	} else if len(args) > 1 {
		command = args[1]
		idx = 2
	}
	return providerArg, command, args[idx:]
}

func parseParams(values []string) (map[string]string, error) {
	params := map[string]string{}
	for _, entry := range values {
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid param: %s", entry)
		}
		params[parts[0]] = parts[1]
	}
	return params, nil
}

func inferProviderName(source string) string {
	if source == "" || source == "-" {
		return ""
	}
	if u, err := url.Parse(source); err == nil && u.Scheme != "" {
		base := filepath.Base(u.Path)
		return strings.TrimSuffix(base, filepath.Ext(base))
	}
	base := filepath.Base(source)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
