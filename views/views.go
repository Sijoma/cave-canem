package views

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	v1 "k8s.io/api/rbac/v1"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed slackMessageTemplates/*
var slackMessageTemplates embed.FS

type message struct {
	Message     string
	ClusterName string `json:"cluster_name"`
	Rb          *v1.RoleBinding
}

func AddRoleBinding(action string, cluster string, binding *v1.RoleBinding) {
	m := message{
		Message:     "Rolebinding " + action,
		ClusterName: cluster,
		Rb:          binding,
	}
	tpl := renderTemplate(slackMessageTemplates, "slackMessageTemplates/RolebindingUpdate.gotmpl", m)

	err := sendSlackNotification(os.Getenv("SLACK_WEBHOOK"), tpl)
	if err != nil {
		panic(err)
		return
	}
}

func sendSlackNotification(webhookUrl string, slackBody bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, webhookUrl, &slackBody)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("non-ok response returned from Slack")
	}
	return nil
}

func renderTemplate(fs fs.FS, file string, args interface{}) bytes.Buffer {
	var tpl bytes.Buffer
	// read the block-kit definition as a go template

	t := template.Must(template.ParseFS(fs, file))
	err := t.Execute(&tpl, args)
	if err != nil {
		log.Fatalln("Could not render view " + err.Error())
	}
	return tpl
}
