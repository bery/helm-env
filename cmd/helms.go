package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CommandFlags struct {
	DryRun  bool
	Install bool
}

// installCmd wraps the helm 'install' command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "wrapper for helm install, decrypting secrets",
	Long: `This command wraps the default helm install command,
	but decrypting any encrypted values file using Barbican. Available
	arguments are the same as for the default command.`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		out, err := wrapHelmCommand("install", Mode, args)
		if err != nil {
			log.Fatalf("%v", string(out))
		}
		fmt.Printf(string(out))
	},
}

// upgradeCmd wraps the helm 'upgrade' command.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "wrapper for helm upgrade, decrypting secrets",
	Long: `This command wraps the default helm upgrade command,
	but decrypting any encrypted values file using Barbican. Available
	arguments are the same as for the default command.`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		out, err := wrapHelmCommand("upgrade", Mode, args)
		if err != nil {
			log.Fatalf("%v", string(out))
		}
		fmt.Printf(string(out))
	},
}

func wrapHelmCommand(cmd string, mode string, args []string) ([]byte, error) {
	var value string
	for _, pair := range os.Environ() {
		variable := strings.Split(pair, "=")
		if strings.HasPrefix(variable[0], Prefix) {
			log.Debugf("Found %s", variable[0])
			if mode == "rename" {
				normalizedKey := normalizeName(strings.TrimPrefix(variable[0], Prefix))
				value = fmt.Sprintf("%s=%s", normalizedKey, variable[1])
				log.Debugf("Setting %s", value)
			} else if mode == "copy" {
				value = strings.TrimPrefix(pair, variable[0]+"=")
				log.Debugf("Setting %s", value)
			}
			args = append(args, []string{"--set", value}...)
		}
	}
	helmArgs, err := getArgs(args)
	if err != nil {
		return []byte{}, err
	}
	fullArgs := append([]string{cmd}, helmArgs...)
	helmCmd := exec.Command("helm", fullArgs...)
	log.Infof("Running helm command: %s", helmCmd.String())
	return helmCmd.CombinedOutput()
}

func getArgs(args []string) ([]string, error) {
	helmArgs := args
	// for i, flag := range args {
	// 	if flag == "--values" || flag == "-f" || flag == "--filename" {
	// 		if len(helmArgs) > i+1 {
	// 			fname := helmArgs[i+1]

	// 			// Update args to access the decrypt shm file instead
	// 			helmArgs[i+1] = tmpf
	// 		}
	// 	}
	// }
	return helmArgs, nil
}

func normalizeName(name string) string {
	for strings.Contains(name, "___") {
		match, _ := GetStringInBetweenTwoString(name, "___", "___")
		replaceForBrackets := fmt.Sprintf("___%s___", match)
		pattern := "[%s]"
		if !strings.HasSuffix(name, replaceForBrackets) {
			pattern = "[%s]."
		}
		name = strings.ReplaceAll(name, replaceForBrackets, fmt.Sprintf(pattern, match))
	}
	return strings.ReplaceAll(name, "__", ".")
}

func GetStringInBetweenTwoString(str string, startS string, endS string) (result string, found bool) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result, false
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result, false
	}
	result = newS[:e]
	return result, true
}
