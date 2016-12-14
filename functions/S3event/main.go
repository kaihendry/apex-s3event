package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apex/go-apex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func getEnvmap() map[string]string {
	envmap := make(map[string]string)
	for _, e := range os.Environ() {
		ep := strings.SplitN(e, "=", 2)
		if ep[0] == "AWS_SECRET_ACCESS_KEY" {
			continue
		}
		if ep[0] == "AWS_SESSION_TOKEN" {
			continue
		}
		envmap[ep[0]] = ep[1]
	}
	return envmap
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {

		// Stringify the incoming event
		var obj interface{}
		json.Unmarshal(event, &obj)
		b, _ := json.MarshalIndent(obj, "", "   ")

		// JSON.stringify the context for the template
		ctxjson, _ := json.MarshalIndent(ctx, "", "   ")

		templates := template.Must(template.New("main").Funcs(template.FuncMap{"time": time.Now}).ParseGlob("templates/*.html"))
		templates = template.Must(templates.ParseGlob("templates/includes/*.html"))

		fn := "/tmp/index.html"
		outputfile, err := os.Create(fn)
		if err != nil {
			panic(err) // Not sure if Panic is the right approach within Lambda
		}

		templates.ExecuteTemplate(outputfile, "index.html", struct {
			Input  string
			Indent string
			Env    map[string]string
			Ctx    string
		}{
			string(event),
			string(b),
			getEnvmap(),
			string(ctxjson),
		})

		s3url, err := url.Parse(os.Getenv("S3URI"))
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stderr, "%#+ v \n", s3url)

		sess, _ := session.NewSession()
		svc := s3.New(sess)

		// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#PutObjectInput
		// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#example_S3_PutObject
		params := &s3.PutObjectInput{
			Bucket:      aws.String(s3url.Host),     // Required
			Body:        outputfile,                 // Required
			Key:         aws.String(s3url.Path[1:]), // Required
			ACL:         aws.String("public-read"),
			ContentType: aws.String("text/html"),
		}

		resp, err := svc.PutObject(params)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to upload", err.Error())
			fmt.Fprintln(os.Stderr, "Does the", os.Getenv("ROLE"), " role in IAM have S3 permissions?")
		} else {
			fmt.Fprintln(os.Stderr, "Managed to upload", resp)
		}

		return "http://" + s3url.Host + ".s3-website-" + os.Getenv("AWS_REGION") + ".amazonaws.com" + s3url.Path, nil
	})
}
