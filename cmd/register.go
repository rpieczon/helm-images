package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	"github.com/nikhilsbhat/helm-images/version"
	"github.com/spf13/cobra"
)

var images = pkg.Images{}

const (
	getArgumentCountLocal   = 2
	getArgumentCountRelease = 1
)

type imagesCommands struct {
	commands []*cobra.Command
}

// SetImagesCommands helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetImagesCommands() *cobra.Command {
	return getImagesCommands()
}

// Add an entry in below function to register new command.
func getImagesCommands() *cobra.Command {
	command := new(imagesCommands)
	command.commands = append(command.commands, getImagesCommand())
	command.commands = append(command.commands, getVersionCommand())

	return command.prepareCommands()
}

func (c *imagesCommands) prepareCommands() *cobra.Command {
	rootCmd := getRootCommand()
	for _, cmnd := range c.commands {
		rootCmd.AddCommand(cmnd)
	}

	registerFlags(rootCmd)

	return rootCmd
}

func getImagesCommand() *cobra.Command {
	imageCommand := &cobra.Command{
		Use:   "get [RELEASE] [CHART] [flags]",
		Short: "Fetches all images those are part of specified chart/release",
		Long:  "Lists all images those are part of specified chart/release and matches the pattern or part of specified registry.",
		Example: `  helm images get prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm images get prometheus-standalone --from-release --registry quay.io
  helm images get prometheus-standalone --from-release --registry quay.io --unique
  helm images get prometheus-standalone --from-release --registry quay.io --yaml`,
		Args: minimumArgError,
		RunE: func(cmd *cobra.Command, args []string) error {
			images.SetLogger(images.LogLevel)
			images.SetWriter(os.Stdout)
			cmd.SilenceUsage = true

			images.SetRelease(args[0])
			if !images.FromRelease {
				images.SetChart(args[1])
			}

			if (images.JSON && images.YAML && images.Table) || (images.JSON && images.YAML) ||
				(images.Table && images.YAML) || (images.Table && images.JSON) {
				return &imgErrors.MultipleFormatError{
					Message: "cannot render the output to multiple format, enable any of '--yaml --json --table' at a time",
				}
			}

			return images.GetImages()
		},
	}
	registerGetFlags(imageCommand)

	return imageCommand
}

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "images [command]",
		Short: "Utility that helps in fetching images which are part of deployment",
		Long:  `Lists all images that would be part of helm deployment.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())

	return rootCommand
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of helm-images installed",
		Long:  `This will help user to find what version of helm-images plugin he/she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func versionConfig(_ *cobra.Command, _ []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		log.Fatalf("fetching version of helm-images failed with: %v", err)
	}

	writer := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"images version", string(buildInfo)}, ": "))

	_, err = writer.WriteString(versionInfo)
	if err != nil {
		log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			log.Fatalln(err)
		}
	}(writer)

	return nil
}

//nolint:goerr113
func minimumArgError(cmd *cobra.Command, args []string) error {
	minArgError := errors.New("[RELEASE] or [CHART] cannot be empty")
	oneOfThemError := errors.New("when '--from-release' is enabled, valid input is [RELEASE] and not both [RELEASE] [CHART]")
	cmd.SilenceUsage = true

	if !images.FromRelease {
		if len(args) != getArgumentCountLocal {
			log.Println(minArgError)

			return minArgError
		}

		return nil
	}

	if len(args) > getArgumentCountRelease {
		log.Fatalln(oneOfThemError)
	}

	return nil
}

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{printf "\n"}}`
}
