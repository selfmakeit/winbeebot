package winbeebot

import (
	"strings"
	"unicode/utf16"
)

type ParsedMessageEntity struct {
	MessageEntity
	Text string `json:"text"`
}

// ParseEntities calls Message.ParseEntity on all message text entities.
func (m Message) ParseEntities() (out []ParsedMessageEntity) {
	return m.ParseEntityTypes(nil)
}

// ParseCaptionEntities calls Message.ParseEntity on all message caption entities.
func (m Message) ParseCaptionEntities() (out []ParsedMessageEntity) {
	return m.ParseCaptionEntityTypes(nil)
}

// ParseEntityTypes calls Message.ParseEntity on a subset of message text entities.
func (m Message) ParseEntityTypes(accepted map[string]struct{}) (out []ParsedMessageEntity) {
	utf16Text := utf16.Encode([]rune(m.Text))
	for _, ent := range m.Entities {
		if _, ok := accepted[ent.Type]; ok || accepted == nil {
			out = append(out, parseEntity(ent, utf16Text))
		}
	}
	return out
}
func (m *Message) IsCommand() bool {
	if m.Entities == nil || len(m.Entities) == 0 {
		return false
	}

	entity := m.Entities[0]
	return entity.Offset == 0 && entity.Type == "bot_command"
}

// Command checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is removed. Use
// CommandWithAt() if you do not want that.
func (m *Message) Command() string {
	command := m.CommandWithAt()

	if i := strings.Index(command, "@"); i != -1 {
		command = command[:i]
	}

	return command
}

// CommandWithAt checks if the message was a command and if it was, returns the
// command. If the Message was not a command, it returns an empty string.
//
// If the command contains the at name syntax, it is not removed. Use Command()
// if you want that.
func (m *Message) CommandWithAt() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := m.Entities[0]
	return m.Text[1:entity.Length]
}

// CommandArguments checks if the message was a command and if it was,
// returns all text after the command name. If the Message was not a
// command, it returns an empty string.
//
// Note: The first character after the command name is omitted:
// - "/foo bar baz" yields "bar baz", not " bar baz"
// - "/foo-bar baz" yields "bar baz", too
// Even though the latter is not a command conforming to the spec, the API
// marks "/foo" as command entity.
func (m *Message) CommandArguments() string {
	if !m.IsCommand() {
		return ""
	}

	// IsCommand() checks that the message begins with a bot_command entity
	entity := m.Entities[0]

	if int64(len(m.Text)) == entity.Length {
		return "" // The command makes up the whole message
	}

	return m.Text[entity.Length+1:]
}

// ParseCaptionEntityTypes calls Message.ParseEntity on a subset of message caption entities.
func (m Message) ParseCaptionEntityTypes(accepted map[string]struct{}) (out []ParsedMessageEntity) {
	utf16Caption := utf16.Encode([]rune(m.Caption))
	for _, ent := range m.CaptionEntities {
		if _, ok := accepted[ent.Type]; ok || accepted == nil {
			out = append(out, parseEntity(ent, utf16Caption))
		}
	}
	return out
}

// ParseEntity parses a single message text entity to populate text contents, URL, and offsets in UTF8.
func (m Message) ParseEntity(entity MessageEntity) ParsedMessageEntity {
	return parseEntity(entity, utf16.Encode([]rune(m.Text)))
}

// ParseCaptionEntity parses a single message caption entity to populate text contents, URL, and offsets in UTF8.
func (m Message) ParseCaptionEntity(entity MessageEntity) ParsedMessageEntity {
	return parseEntity(entity, utf16.Encode([]rune(m.Caption)))
}

func parseEntity(entity MessageEntity, utf16Text []uint16) ParsedMessageEntity {
	text := string(utf16.Decode(utf16Text[entity.Offset : entity.Offset+entity.Length]))

	if entity.Type == "url" {
		entity.Url = text
	}

	entity.Offset = int64(len(string(utf16.Decode(utf16Text[:entity.Offset]))))
	entity.Length = int64(len(text))

	return ParsedMessageEntity{
		MessageEntity: entity,
		Text:          text,
	}
}
