package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"flag"

	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

func main() {
	command := flag.String("command", "pwd", "Command to run")

	flag.Parse()

	config := getConfig()
	u, err := getUser(config.orgToken) // TODO: change to personal user token instead of orgToken
	if err != nil {
		log.Fatal(err)
	}

	isMember, err := checkOrgMembership(*u.Login, config.organization, config.orgToken)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("%s is member of %s %t", *u.Login, config.organization, isMember)

	if isMember{
		output, err := remoteRun("niko","dev-cat-web-platform-0", config.sshKey, *command)
		if err != nil{
			log.Fatal(err)
		}
		log.Println(output)
	}
}

func remoteRun(user string, addr string, sshKey string, cmd string) (string, error) {
	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring
	key, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return "", err
	}
	// Authentication
	config := &ssh.ClientConfig{
		User: user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),	
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}
	// Connect
	client, err := ssh.Dial("tcp", addr+":22", config)
	if err != nil {
		return "", err
	}
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer  // import "bytes"
	session.Stdout = &b // get output
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")
	// Finally, run the command
	err = session.Run(cmd)
	return b.String(), err
}

type Config struct {
	organization string
	sshKey       string
	orgToken     string
}

func getConfig() Config {
	orgToken, ok := os.LookupEnv("GITHUBTOKEN") // sshUsersGithubToken
	if !ok {
		log.Fatalf("Missing GITHUBTOKEN environment variable")
	}

	organization, ok := os.LookupEnv("GITHUBORG")
	if !ok {
		log.Fatalf("Missing GITHUBORGA environment variable")
	}

	sshKey, ok := os.LookupEnv("SSHKEY")
	if !ok {
		log.Fatalf("Missing SSHKEY environment variable")
	}

	return Config{
		sshKey:       sshKey,
		organization: organization,
		orgToken:     orgToken}
}

func checkOrgMembership(userLogin, orgLogin, appToken string) (bool, error) {
	ctx := context.Background()
	oauthTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: appToken},
	)
	oauthTokenClient := oauth2.NewClient(ctx, oauthTokenSource)
	githubClient := github.NewClient(oauthTokenClient)
	isMember, _, err := githubClient.Organizations.IsMember(ctx, orgLogin, userLogin)

	return isMember, err

}

func getUser(userToken string) (*github.User, error) {
	ctx := context.Background()
	oauthTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: userToken},
	)
	oauthTokenClient := oauth2.NewClient(ctx, oauthTokenSource)
	githubClient := github.NewClient(oauthTokenClient)
	u, _, err := githubClient.Users.Get(ctx, "")
	return u, err
}
