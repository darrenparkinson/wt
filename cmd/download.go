package cmd

import (
	"fmt"
	"log"

	"github.com/darrenparkinson/wt/internal/webex"
	"github.com/spf13/cobra"
)

var (
	nbrDomain   string
	recordingID string
	dlyear      int
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download WebEx Recordings",
	Long: `
The download command enables IT Administrators to download WebEx Recordings.`,
	Run: func(cmd *cobra.Command, args []string) {
		if tenant == "" || siteID == "" || nbrDomain == "" {
			log.Fatalln("tenant name, site ID and nbr domain are required")
		}
		if username == "" || password == "" {
			username, password = credentials()
		}
		r, err := webex.DownloadRecording(username, password, tenant, siteID, nbrDomain, recordingID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)

	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&recordingID, "recid", "r", "", "recordingid of recording to download (required)")
	downloadCmd.PersistentFlags().StringVarP(&username, "username", "", "", "admin username, i.e. bjones")
	downloadCmd.PersistentFlags().StringVarP(&password, "password", "", "", "admin password")
	downloadCmd.PersistentFlags().StringVarP(&tenant, "tenant", "", "", "tenant name of your webex tenant without .webex.com, e.g. acme")
	downloadCmd.PersistentFlags().StringVarP(&siteID, "site", "", "", "site ID of your webex tenant, typically a six digit number")
	downloadCmd.PersistentFlags().StringVarP(&nbrDomain, "domain", "", "", "domain of your nbr dc, without .webex.com")

	downloadCmd.MarkFlagRequired("recid")

}
