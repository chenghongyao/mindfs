package acp

import (
	"bufio"
	"encoding/json"
	"io"
)

// Parser reads ndJSON messages from a stream.
type Parser struct {
	reader *bufio.Reader
}

// NewParser creates a new ndJSON parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

// ReadMessage reads and parses the next JSON message.
// Returns io.EOF when the stream ends.
func (p *Parser) ReadMessage() (ACPMessage, error) {
	line, err := p.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(line) > 0 {
			// Handle final line without newline
			var msg ACPMessage
			if jsonErr := json.Unmarshal(line, &msg); jsonErr != nil {
				return ACPMessage{}, jsonErr
			}
			return msg, io.EOF
		}
		return ACPMessage{}, err
	}

	// Skip empty lines
	if len(line) == 0 || (len(line) == 1 && line[0] == '\n') {
		return p.ReadMessage()
	}

	var msg ACPMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return ACPMessage{}, err
	}
	return msg, nil
}

// ReadAll reads all messages until EOF.
func (p *Parser) ReadAll() ([]ACPMessage, error) {
	var messages []ACPMessage
	for {
		msg, err := p.ReadMessage()
		if err == io.EOF {
			if msg.Method != "" || msg.Result != nil {
				messages = append(messages, msg)
			}
			return messages, nil
		}
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}
}
