package main

import (
    // "encoding/json"
    "fmt"
    "os"
    "strings"
    "path/filepath"
    "time"
    "strconv"

    "github.com/aws/aws-sdk-go/aws"
    // "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"

)


type (
    S3Event struct {
        Records []struct {
            EventVersion string    `json:"eventVersion"`
            EventSource  string    `json:"eventSource"`
            AwsRegion    string    `json:"awsRegion"`
            EventTime    time.Time `json:"eventTime"`
            EventName    string    `json:"eventName"`
            UserIdentity struct {
                PrincipalID string `json:"principalId"`
            } `json:"userIdentity"`
            RequestParameters struct {
                SourceIPAddress string `json:"sourceIPAddress"`
            } `json:"requestParameters"`
            ResponseElements struct {
                XAmzRequestID string `json:"x-amz-request-id"`
                XAmzID2       string `json:"x-amz-id-2"`
            } `json:"responseElements"`
            S3 struct {
                S3SchemaVersion string `json:"s3SchemaVersion"`
                ConfigurationID string `json:"configurationId"`
                Bucket          struct {
                    Name          string `json:"name"`
                    OwnerIdentity struct {
                        PrincipalID string `json:"principalId"`
                    } `json:"ownerIdentity"`
                    Arn string `json:"arn"`
                } `json:"bucket"`
                Object struct {
                    Key       string `json:"key"`
                    Size      int    `json:"size"`
                    ETag      string `json:"eTag"`
                    Sequencer string `json:"sequencer"`
                } `json:"object"`
            } `json:"s3"`
        } `json:"Records"`
    }
)

func handler(s3Json S3Event) error {
    var debug bool
    var AwsID, AwsKey string
    // var s3Json s3Levent

    debug = false
    debugTxt := os.Getenv("Debug")
    AwsID  = os.Getenv("AKID")
    AwsKey = os.Getenv("SECRETKEY")

    if debugTxt == "TRUE" {
        debug = true
    }

    if debug {
        //fmt.Println("S3Event %o", s3Json)
    }

    sess, err := session.NewSession(&aws.Config{
        Credentials: credentials.NewStaticCredentials(AwsID, AwsKey, ""),
    })

    bucketName := s3Json.Records[0].S3.Bucket.Name
    newFile := s3Json.Records[0].S3.Object
    folder := filepath.Dir(newFile.Key)

    if debug {
        fmt.Println("Event Bucket %s, Folder %s, File : %v", bucketName, folder, newFile)
    }

    svc := s3.New(sess)
    input := &s3.ListObjectsV2Input{
        Bucket: aws.String(bucketName),
        Prefix: aws.String(folder),
        MaxKeys: aws.Int64(2),
    }

    result, err := svc.ListObjectsV2(input)
    if err != nil {
        if debug {
            fmt.Println("ListObjectsV2 Error %o", err)
        }
        return nil
    }

    html := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html><head>
    <title>Index of {Folder}</title></head><body><h1>Index of {Folder}</h1><table>
    <tr><th valign="top"><img src="/icons/blank.gif" ></th><th><a href="?C=N;O=D">Name</a></th><th>
    <a href="?C=M;O=A">Last modified</a></th><th><a href="?C=S;O=A">Size</a></th><th><a href="?C=D;O=A">Description</a></th></tr>
    <tr><th colspan="5"><hr></th></tr><tr><td valign="top"><img src="/icons/back.gif" ></td><td><a href="/download/apps/">Parent Directory</a></td><td>&nbsp;</td><td align="right">  - </td><td>&nbsp;</td></tr>`

    tr := `<tr><td valign="top"><img src="/icons/unknown.gif" ></td><td><a href="{NodeName}">{NodeName}</a></td><td align="right">{NodeTime}</td><td align="right">{NodeSize}</td><td>&nbsp;</td></tr>`

    html = strings.Replace(html, "{Folder}", folder, -1)

    var tr2 string
    for _, item := range result.Contents {
        tr2 = strings.Replace(tr, "{NodeName}", filepath.Base(*item.Key), -1)
        tr2 = strings.Replace(tr2, "{NodeTime}", item.LastModified.Format("2006-01-02T15:04:05"), -1)
        tr2 = strings.Replace(tr2, "{NodeSize}", strconv.FormatInt(*item.Size, 10), -1)
        html = html + tr2

        if debug {
            fmt.Println("Name:   %s ;;  %s     ", *item.Key, filepath.Base(*item.Key))
            fmt.Println("Last modified:", *item.LastModified)
            fmt.Println("Size:         ", *item.Size)
            fmt.Println("Storage class:", *item.StorageClass)
            fmt.Println("")
        }
    }

    t := time.Now()

    tr2 = strings.Replace(tr, "{NodeName}", filepath.Base(newFile.Key), -1)
    tr2 = strings.Replace(tr2, "{NodeTime}", t.Format("2006-01-02T15:04:05"), -1)
    tr2 = strings.Replace(tr2, "{NodeSize}", strconv.Itoa(newFile.Size), -1)
    html = html + tr2 + `<tr><th colspan="5"><hr></th></tr></table></body></html>`

    input2 := &s3.PutObjectInput{
        Body:   aws.ReadSeekCloser(strings.NewReader(html)),
        Bucket: aws.String(bucketName),
        Key:    aws.String(folder + "/index.html"),
    }

    _, err9 := svc.PutObject(input2)
    if err9 != nil {
        fmt.Println(err9.Error())
        return nil
    }

    return nil
}

func main() {
    lambda.Start(handler)
}