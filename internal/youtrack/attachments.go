package youtrack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

const attachmentFields = "id,name,url,size,mimeType,created"

func (c *Client) ListAttachments(id string) ([]Attachment, error) {
	params := url.Values{"fields": {attachmentFields}}

	data, err := c.get(entityPath(id)+"/attachments", params)
	if err != nil {
		return nil, fmt.Errorf("list attachments for %s: %w", id, err)
	}

	var attachments []Attachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("parse attachments: %w", err)
	}

	return attachments, nil
}

func (c *Client) DownloadAttachment(relURL string, w io.Writer) error {
	return c.download(relURL, w)
}

// UploadFile is one file to attach to an issue or article.
type UploadFile struct {
	Name    string
	Content io.Reader
}

// UploadAttachments attaches files to an issue or article in one request and
// returns the created attachments.
func (c *Client) UploadAttachments(id string, files []UploadFile) ([]Attachment, error) {
	path := entityPath(id) + "/attachments?fields=" + url.QueryEscape(attachmentFields)
	data, err := c.postMultipart(path, files)
	if err != nil {
		return nil, fmt.Errorf("upload attachments to %s: %w", id, err)
	}

	var attachments []Attachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("parse uploaded attachments: %w", err)
	}
	return attachments, nil
}
