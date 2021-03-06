package courier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	validator "gopkg.in/go-playground/validator.v9"
)

// WriteError writes a JSON response for the passed in error
func WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	errors := []string{err.Error()}

	vErrs, isValidation := err.(validator.ValidationErrors)
	if isValidation {
		errors = []string{}
		for i := range vErrs {
			errors = append(errors, fmt.Sprintf("field '%s' %s", strings.ToLower(vErrs[i].Field()), vErrs[i].Tag()))
		}
	}
	return writeJSONResponse(w, http.StatusBadRequest, &errorResponse{errors})
}

// WriteIgnored writes a JSON response for the passed in message
func WriteIgnored(w http.ResponseWriter, r *http.Request, details string) error {
	LogRequestIgnored(r, details)
	return writeData(w, http.StatusOK, details, struct{}{})
}

// WriteChannelEventSuccess writes a JSON response for the passed in event indicating we handled it
func WriteChannelEventSuccess(w http.ResponseWriter, r *http.Request, event ChannelEvent) error {
	LogChannelEventReceived(r, event)
	return writeData(w, http.StatusOK, "Event Accepted",
		&eventReceiveData{
			event.ChannelUUID(),
			event.EventType(),
			event.URN(),
			event.CreatedOn(),
		})
}

// WriteMsgSuccess writes a JSON response for the passed in msg indicating we handled it
func WriteMsgSuccess(w http.ResponseWriter, r *http.Request, msg Msg) error {
	LogMsgReceived(r, msg)
	return writeData(w, http.StatusOK, "Message Accepted",
		&msgReceiveData{
			msg.Channel().UUID(),
			msg.UUID(),
			msg.Text(),
			msg.URN(),
			msg.Attachments(),
			msg.ExternalID(),
			msg.ReceivedOn(),
		})
}

// WriteStatusSuccess writes a JSON response for the passed in status update indicating we handled it
func WriteStatusSuccess(w http.ResponseWriter, r *http.Request, status MsgStatus) error {
	LogMsgStatusReceived(r, status)
	return writeData(w, http.StatusOK, "Status Update Accepted",
		&statusData{
			status.ChannelUUID(),
			status.Status(),
			status.ID(),
			status.ExternalID(),
		})
}

type errorResponse struct {
	Errors []string `json:"errors"`
}

type successResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type msgReceiveData struct {
	ChannelUUID ChannelUUID `json:"channel_uuid"`
	MsgUUID     MsgUUID     `json:"msg_uuid"`
	Text        string      `json:"text"`
	URN         URN         `json:"urn"`
	Attachments []string    `json:"attachments,omitempty"`
	ExternalID  string      `json:"external_id,omitempty"`
	ReceivedOn  *time.Time  `json:"received_on,omitempty"`
}

type eventReceiveData struct {
	ChannelUUID ChannelUUID      `json:"channel_uuid"`
	EventType   ChannelEventType `json:"event_type"`
	URN         URN              `json:"urn"`
	ReceivedOn  time.Time        `json:"received_on"`
}

type statusData struct {
	ChannelUUID ChannelUUID    `json:"channel_uuid"`
	Status      MsgStatusValue `json:"status"`
	MsgID       MsgID          `json:"msg_id,omitempty"`
	ExternalID  string         `json:"external_id,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}

func writeData(w http.ResponseWriter, statusCode int, message string, response interface{}) error {
	return writeJSONResponse(w, statusCode, &successResponse{message, response})
}
