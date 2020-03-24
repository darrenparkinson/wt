# WT - Webex Tool

> Note: This tool has been uploaded on request.  It is some of my early Go code so please don't judge me 😬😉.  

>It was intended as a management utility for our webex tenant, but we have since switched over to a new SSO based tenant where this no longer works.  You can use the DevNet sandbox for testing however.  

>This code accompanies the "[Downloading Webex Recordings with Go](http://darrenparkinson.uk/posts/2019-05-04-downloading-webex-recordings-with-go/)" blog post.  

## Installation

To get the latest version of this tool:

```
go get github.com/darrenparkinson/wt
```

## Command Help

To get help with a command, you can just type `wt help` or `wt help list` or `wt help download` which will then provide you with a list of available parameters etc.

```
WebEx Tool (wt) is a tool for IT Administrators to manage their WebEx Tenant.

Usage:
wt [command]

Available Commands:
  download    Download WebEx Recordings
  help        Help about any command
  list        List WebEx Recordings
  version     Print the version number of wt

Flags:
-h, --help help for wt

Use "wt [command] --help" for more information about a command.
```

To list recordings for example you might type something like:

```
$ wt list --tenant apidemoeu --site 678910 --username myusername --password mypassword --year 2020
```

OR if you're an admin, you can list recordings for another user:

```
$ wt list --tenant apidemoeu --site 678910 --username myusername --password mypassword --userid someotherusername

Host WebExID         RecordingID      Size                 CreateTime     Name
-----------------    -----------      ----                 ----------     ----
someotherusername      181204076     2.092       03/24/2020 17:00:39     my_test_meeting-20200324 0900-1
```

If you use the `--csv` option you will get a few extra details, including the file URL.

To download recordings, rather annoyingly you actually have to specify the data center specific domain for your recordings in addition to the other details required for listing.  As per [the documentation](https://developer.cisco.com/docs/webex-meetings/#!nbr-web-services-api) you will need to obtain this from Cisco. 

```
$ wt download --tenant apidemoeu --site 678910 --domain nln1wss1 --username myusername --password mypassword --recid 181204076
```

If you omit the username or password on the command line (which you should), you will be prompted for them.

## Sandbox

To test out the command, you can use the [DevNet Sandbox for Webex](https://developer.cisco.com/site/sandbox/) which provides a fully functioning Webex Tenant.  

The WebEx Lab details provide the information required to use it such as the tenant name (e.g. apidemoeu) and site ID.

The only caveat is that to get a recording you have to use Webex Training and you will need to get the NBR domain from Cisco (I'll update here if I find it out for the sandbox).

