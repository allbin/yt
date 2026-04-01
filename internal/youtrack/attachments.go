package youtrack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

const attachmentFields = "id,name,url,size,mimeType,created"

func (c *Client) ListAttachments(issueID string) ([]Attachment, error) {
	params := url.Values{"fields": {attachmentFields}}

	data, err := c.get("/api/issues/"+url.PathEscape(issueID)+"/attachments", params)
	if err != nil {
		return nil, fmt.Errorf("list attachments for %s: %w", issueID, err)
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
