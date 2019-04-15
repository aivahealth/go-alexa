package skillserver

import (
	"encoding/json"
	"errors"
	"time"
)

// Request Functions
func (this *EchoRequest) VerifyTimestamp() bool {
	reqTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", this.Request.Timestamp)
	if time.Since(reqTimestamp) < time.Duration(150)*time.Second {
		return true
	}

	return false
}

func (this *EchoRequest) VerifyAppID(myAppID string) bool {
	if this.Session.Application.ApplicationID == myAppID ||
		this.Context.System.Application.ApplicationID == myAppID {
		return true
	}

	return false
}

func (this *EchoRequest) GetSessionID() string {
	return this.Session.SessionID
}

func (this *EchoRequest) GetUserID() string {
	// If there's no session, the userid is present in the context/system object.
	uid := this.Session.User.UserID
	if uid == "" {
		uid = this.Context.System.User.UserId
	}
	return uid
}

func (this *EchoRequest) GetRequestType() string {
	return this.Request.Type
}

func (this *EchoRequest) GetIntentName() string {
	if this.GetRequestType() == "IntentRequest" {
		return this.Request.Intent.Name
	}

	return this.GetRequestType()
}

func (this *EchoRequest) GetSlotValue(slotName string) (string, error) {
	if _, ok := this.Request.Intent.Slots[slotName]; ok {
		return this.Request.Intent.Slots[slotName].Value, nil
	}

	return "", errors.New("Slot name not found.")
}

func (this *EchoRequest) AllSlots() map[string]EchoSlot {
	return this.Request.Intent.Slots
}

// Response Functions
func NewEchoResponse() *EchoResponse {
	trueBool := true
	er := &EchoResponse{
		Version: "1.0",
		Response: EchoRespBody{
			ShouldEndSession: &trueBool,
			Directives:       []Directive{},
		},
		SessionAttributes: make(map[string]interface{}),
	}

	return er
}

func (this *EchoResponse) OutputSpeech(text string) *EchoResponse {
	this.Response.OutputSpeech = &EchoRespPayload{
		Type: "PlainText",
		Text: text,
	}

	return this
}

func (this *EchoResponse) Card(title string, content string) *EchoResponse {
	return this.SimpleCard(title, content)
}

func (this *EchoResponse) OutputSpeechSSML(text string) *EchoResponse {
	this.Response.OutputSpeech = &EchoRespPayload{
		Type: "SSML",
		SSML: text,
	}

	return this
}

func (this *EchoResponse) SimpleCard(title string, content string) *EchoResponse {
	this.Response.Card = &EchoRespPayload{
		Type:    "Simple",
		Title:   title,
		Content: content,
	}

	return this
}

func (this *EchoResponse) StandardCard(title string, content string, smallImg string, largeImg string) *EchoResponse {
	this.Response.Card = &EchoRespPayload{
		Type:    "Standard",
		Title:   title,
		Content: content,
	}

	if smallImg != "" {
		this.Response.Card.Image.SmallImageURL = smallImg
	}

	if largeImg != "" {
		this.Response.Card.Image.LargeImageURL = largeImg
	}

	return this
}

func (this *EchoResponse) LinkAccountCard() *EchoResponse {
	this.Response.Card = &EchoRespPayload{
		Type: "LinkAccount",
	}

	return this
}

func (this *EchoResponse) Reprompt(text string) *EchoResponse {
	this.Response.Reprompt = &EchoReprompt{
		OutputSpeech: EchoRespPayload{
			Type: "PlainText",
			Text: text,
		},
	}

	return this
}

func (this *EchoResponse) RepromptSSML(text string) *EchoResponse {
	this.Response.Reprompt = &EchoReprompt{
		OutputSpeech: EchoRespPayload{
			Type: "SSML",
			Text: text,
		},
	}

	return this
}

func (this *EchoResponse) EndSession(flag bool) *EchoResponse {
	this.Response.ShouldEndSession = &flag

	return this
}

func (this *EchoResponse) String() ([]byte, error) {
	jsonStr, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}

// AudioPlayer interface

type AudioPlayerPlayBehavior string

const (
	ReplaceAll      AudioPlayerPlayBehavior = "REPLACE_ALL"
	Enqueue         AudioPlayerPlayBehavior = "ENQUEUE"
	ReplaceEnqueued AudioPlayerPlayBehavior = "REPLACE_ENQUEUED"
)

func (this *EchoResponse) AudioPlayerPlay(
	behavior AudioPlayerPlayBehavior, streamUrl, token string, prevToken *string, offsetMs int,
) *EchoResponse {
	streamObj := map[string]interface{}{
		"url":                  streamUrl,
		"token":                token,
		"offsetInMilliseconds": offsetMs,
	}
	if prevToken != nil {
		streamObj["expectedPreviousToken"] = *prevToken
	}
	directive := map[string]interface{}{
		"type":         "AudioPlayer.Play",
		"playBehavior": behavior,
		"audioItem": map[string]interface{}{
			"stream": streamObj,
		},
	}
	this.Response.Directives = append(this.Response.Directives, directive)
	return this
}

func (this *EchoResponse) AudioPlayerStop() *EchoResponse {
	directive := map[string]interface{}{
		"type": "AudioPlayer.Stop",
	}
	this.Response.Directives = append(this.Response.Directives, directive)
	return this
}

type AudioPlayerClearQueueBehavior string

const (
	ClearEnqueued AudioPlayerClearQueueBehavior = "CLEAR_ENQUEUED"
	ClearAll      AudioPlayerClearQueueBehavior = "CLEAR_ALL"
)

func (this *EchoResponse) AudioPlayerClearQueue(clearBehavior AudioPlayerClearQueueBehavior) *EchoResponse {
	directive := map[string]interface{}{
		"type":          "AudioPlayer.ClearQueue",
		"clearBehavior": clearBehavior,
	}
	this.Response.Directives = append(this.Response.Directives, directive)
	return this
}

// VideoApp interface

func (this *EchoResponse) VideoAppLaunch(
	streamUrl, title, subtitle string,
) *EchoResponse {
	videoItemObj := map[string]interface{}{
		"source": streamUrl,
	}
	if title != "" || subtitle != "" {
		videoItemObj["metadata"] = map[string]interface{}{
			"title":    title,
			"subtitle": subtitle,
		}
	}
	directive := map[string]interface{}{
		"type":      "VideoApp.Launch",
		"videoItem": videoItemObj,
	}
	this.Response.Directives = append(this.Response.Directives, directive)
	return this
}

// Request Types

type EchoRequest struct {
	Version string      `json:"version"`
	Session EchoSession `json:"session"`
	Request EchoReqBody `json:"request"`
	Context EchoContext `json:"context"`
}

type EchoSession struct {
	New         bool   `json:"new"`
	SessionID   string `json:"sessionId"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
	Attributes map[string]interface{} `json:"attributes"`
	User       struct {
		UserID      string `json:"userId"`
		AccessToken string `json:"accessToken,omitempty"`
	} `json:"user"`
}

type EchoContext struct {
	System struct {
		ApiEndpoint    string `json:"apiEndpoint,omitempty"`
		ApiAccessToken string `json:"apiAccessToken,omitempty"`
		Device         struct {
			DeviceId string `json:"deviceId,omitempty"`
		} `json:"device,omitempty"`
		Application struct {
			ApplicationID string `json:"applicationId,omitempty"`
		} `json:"application,omitempty"`
		User struct {
			AccessToken string `json:"accessToken,omitempty"`
			UserId      string `json:"userId,omitempty"`
			Permissions struct {
				ConsentToken string `json:"consentToken,omitempty"`
			} `json:"permissions,omitempty"`
		} `json:"user,omitempty"`
	} `json:"System,omitempty"`
}

type EchoReqBody struct {
	Type      string            `json:"type"`
	RequestID string            `json:"requestId"`
	Timestamp string            `json:"timestamp"`
	Intent    EchoIntent        `json:"intent,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	Message   map[string]string `json:"message"`
	Locale    string            `json:"locale"`
}

type EchoIntent struct {
	Name  string              `json:"name"`
	Slots map[string]EchoSlot `json:"slots"`
}

type EchoSlot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Response Types

type EchoResponse struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          EchoRespBody           `json:"response"`
}

type EchoRespBody struct {
	OutputSpeech     *EchoRespPayload         `json:"outputSpeech,omitempty"`
	Card             *EchoRespPayload         `json:"card,omitempty"`
	Reprompt         *EchoReprompt            `json:"reprompt,omitempty"`         // Pointer so it's dropped if empty in JSON response.
	ShouldEndSession *bool                    `json:"shouldEndSession,omitempty"` // Same
	Directives       []Directive              `json:"directives"`
	CanFulfillIntent *CanFulfillIntentPayload `json:"canFulfillIntent,omitempty"`
}

type CanFulfillIntentPayload struct {
	CanFulfill CanFulfillIntentAnswer          `json:"canFulfill"`
	Slots      map[string]CanFulfillIntentSlot `json:"slots,omitempty"`
}

type CanFulfillIntentSlot struct {
	CanUnderstand CanFulfillIntentAnswer `json:"canUnderstand"`
	CanFulfill    CanFulfillIntentAnswer `json:"canFulfill"`
}

type CanFulfillIntentAnswer string

const (
	CanFulfillIntentAnswerYes   CanFulfillIntentAnswer = "YES"
	CanFulfillIntentAnswerNo                           = "NO"
	CanFulfillIntentAnswerMaybe                        = "MAYBE"
)

type Directive map[string]interface{} // Shape differs wildly

type EchoReprompt struct {
	OutputSpeech EchoRespPayload `json:"outputSpeech,omitempty"`
}

type EchoRespImage struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

type EchoRespPayload struct {
	Type    string        `json:"type,omitempty"`
	Title   string        `json:"title,omitempty"`
	Text    string        `json:"text,omitempty"`
	SSML    string        `json:"ssml,omitempty"`
	Content string        `json:"content,omitempty"`
	Image   EchoRespImage `json:"image,omitempty"`
}
