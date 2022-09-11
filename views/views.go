package views

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/rbac/v1"
)

//go:embed slackMessageTemplates/*
var slackMessageTemplates embed.FS

type message struct {
	Message     string
	ClusterName string `json:"cluster_name"`
	Rb          *v1.RoleBinding
}

// AddRoleBinding renders a slack message template and sends the notification
func AddRoleBinding(action string, cluster string, binding *v1.RoleBinding) {
	m := message{
		Message:     "Rolebinding " + action,
		ClusterName: cluster,
		Rb:          binding,
	}
	tpl := renderTemplate(slackMessageTemplates, "slackMessageTemplates/RolebindingUpdate.gotmpl", m)

	err := sendSlackNotification(os.Getenv("SLACK_WEBHOOK"), tpl)
	if err != nil {
		log.Printf("error sending slack notification err: %s", err.Error())
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
		log.Printf("Could not render view err: %s", err.Error())
	}
	return tpl
}
