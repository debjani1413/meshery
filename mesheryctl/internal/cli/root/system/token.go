package system

import (
	"fmt"
	"os"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ctx string
var viewAllTokens bool
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage Meshery user tokens",
	Long: `
	Manipulate user tokens and their context assignments in your meshconfig`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok := utils.IsValidSubcommand(availableSubcommands, args[0]); !ok {
			return errors.New(utils.SystemError(fmt.Sprintf("invalid command: \"%s\"", args[0])))
		}
		return nil
	},
}

var createTokenCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a token in your meshconfig",
	Long:  "Create the token with provided token name (optionally token path) to your meshconfig tokens.",
	Example: `
	mesheryctl system token add <token-name> -f <token-path>
	mesheryctl system token add <token-name> (default path is auth.json)
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]
		if tokenPath == "" {
			tokenPath = "auth.json"
		}

		token := config.Token{
			Name:     tokenName,
			Location: tokenPath,
		}
		if err := config.AddTokenToConfig(token, utils.DefaultConfigPath); err != nil {
			return errors.Wrap(err, "Could not create specified token to config")
		}
		log.Printf("Token %s created.", tokenName)
		return nil
	},
}
var deleteTokenCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a token from your meshconfig",
	Long:  "Delete the token with provided token name from your meshconfig tokens.",
	Example: `
	mesheryctl system token delete <token-name>
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]

		if err = config.DeleteTokenFromConfig(tokenName, utils.DefaultConfigPath); err != nil {
			return errors.Wrapf(err, "Could not delete token \"%s\" from config", tokenName)
		}
		log.Printf("Token %s deleted.", tokenName)
		return nil
	},
}
var setTokenCmd = &cobra.Command{
	Use:   "set",
	Short: "Set token for context",
	Long:  "Set token for current context or context specified with --context flag.",
	Example: `
	mesheryctl system token set <token-name> 

	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]
		if ctx == "" {
			ctx = viper.GetString("current-context")

		}

		if err = config.SetTokenToConfig(tokenName, utils.DefaultConfigPath, ctx); err != nil {
			return errors.Wrapf(err, "Could not set token \"%s\" on context %s", tokenName, ctx)

		}
		log.Printf("Token %s set for context %s", tokenName, ctx)
		return nil
	},
}
var listTokenCmd = &cobra.Command{
	Use:   "list",
	Short: "List tokens",
	Long:  "List all the tokens in your meshconfig",
	Example: `
	mesheryctl system token list
	`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(utils.DefaultConfigPath); os.IsNotExist(err) {
			return err
		}

		viper.SetConfigFile(utils.DefaultConfigPath)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}

		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			return errors.Wrap(err, "error processing config")
		}
		log.Print("Available tokens: ")
		for _, t := range *mctlCfg.GetTokens() {
			log.Info(t.Name)
		}
		return nil
	},
}
var viewTokenCmd = &cobra.Command{
	Use:   "view",
	Short: "View token",
	Long:  "View a specific token in meshery config",
	Example: `
	mesheryctl system token view <token-name>
	mesheryctl system token view (show token of current context)
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(utils.DefaultConfigPath); os.IsNotExist(err) {
			return err
		}

		viper.SetConfigFile(utils.DefaultConfigPath)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}

		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			return errors.Wrap(err, "error processing config")
		}
		if viewAllTokens {
			log.Info("Listing all available tokens...\n")
			for _, t := range *mctlCfg.GetTokens() {
				log.Info("- token: ", t.Name)
				log.Info("  location: ", t.Location)
			}
			return nil
		}
		tokenName := ""
		if len(args) == 0 {
			token, err := mctlCfg.GetTokenForContext(viper.GetString("current-context"))
			if err != nil {
				return errors.Wrap(err, "Could not get token for the current context")
			}
			log.Warnf("Token unspecified. Displaying token for current context \"%s\"\n", viper.GetString("current-context"))
			log.Info("token: ", token.Name)
			log.Info("location: ", token.Location)
			return nil
		}
		tokenName = args[0]
		var tokenNames []string
		for _, t := range mctlCfg.Tokens {
			if t.Name == tokenName {
				log.Info("token: ", t.Name)
				log.Info("location: ", t.Location)
				return nil
			}
			// Collecting token names in case the provided token name is invalid and a prompt is to be shown.
			tokenNames = append(tokenNames, t.Name)
		}

		log.Info("Invalid token name. Select from available tokens-")
		prompt := promptui.Select{
			Label: "Select a token from the list",
			Items: tokenNames,
		}
		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		log.Info("token: ", mctlCfg.Tokens[i].Name)
		log.Info("location: ", mctlCfg.Tokens[i].Location)
		return nil
	},
}

func init() {
	tokenCmd.AddCommand(createTokenCmd, deleteTokenCmd, setTokenCmd, listTokenCmd, viewTokenCmd)
	createTokenCmd.Flags().StringVarP(&tokenPath, "filepath", "f", "", "Add the token location")
	setTokenCmd.Flags().StringVar(&ctx, "context", "", "Pass the context")
	viewTokenCmd.Flags().BoolVar(&viewAllTokens, "all", false, "set the flag to view all the tokens.")
}
