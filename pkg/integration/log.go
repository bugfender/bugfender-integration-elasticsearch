package integration

import (
	"time"

	"github.com/gofrs/uuid"
)

//Log log returned by the Bugfender API
type Log struct {
	Uuid uuid.UUID `json:"uuid"`
	// App ID
	App int64 `json:"app"`
	// Device UDID (usually a UUID)
	DeviceUDID string `json:"device.udid"`
	// Device name
	DeviceName string `json:"device.name"`
	// Device model
	DeviceType string `json:"device.type"`
	// App version (eg. 1.2.3.4)
	VersionVersion string `json:"version.version"`
	// App build (eg. 1234)
	VersionBuild string `json:"version.build"`

	// Device language setting
	Language string `json:"language"`
	// Device operating system version
	OSVersion string `json:"os_version"`
	// Device time zone
	Timezone string `json:"timezone"`

	// Log message
	Text string `json:"text"`
	// Method where the log originated
	Method string `json:"method"`
	// File name where the log originated
	File string `json:"file"`
	// Line number where the log originated
	Line int64 `json:"line"`
	// Log level: Fatal (5), Error (2), Warning (1), Info (4), Debug (0), Trace (3) (from more to less critical).
	// Note enum numbers are not sorted for backwards compatibility.
	Level int `json:"log_level"`
	// Log tag
	Tag string `json:"tag"`
	// Timestamp when the log was generated (using the originating device's clock).
	// This timestamp is corrected by Bugfender if it's obviously wrong (far in the past or in the future).
	Time time.Time `json:"time"`
	// Thread ID where the log originated
	ThreadID string `json:"thread_id"`
	// Thread name where the log originated
	ThreadName string `json:"thread_name"`
	// Sorter: specifies the order of occurrence of the logs when multiple logs have the same timestamp (`time` field).
	// For the same `time` value, it is guaranteed no two logs will have the same `absolute_time`.
	// There are no restrictions for this value for different `time`.
	// Therefore, the following tuple is guaranteed to be unique: (`app`, `device.udid`, `time`, `absolute_time`).
	AbsoluteTime uint64 `json:"absolute_time"`
	// URL where the event happened
	URL string `json:"url"`

	// Type of the Log.
	Type string `json:"type"`
	// ID of the Issue associated with the Log, if any.
	IssueID *string `json:"issue_id,omitempty"`
	// If associated Issue exists, contains its' title.
	IssueTitle *string `json:"issue_title,omitempty"`
	// If associated Issue exists, contains its' message.
	IssueMarkdown *string `json:"issue_markdown,omitempty"`
	// If associated Issue exists, contains its' status.
	IssueStatus *int `json:"issue_status,omitempty"`
	// If present, contains name of the Android Activity inside which the Log occurred.
	ActivityName *string `json:"activity_name,omitempty"`
	// If present, contains status of the Android Activity inside which the Log occurred.
	ActivityStatus *string `json:"activity_status,omitempty"`
	// If present, contains name of the iOS VC in which the Log occured.
	ViewControllerName *string `json:"view_controller_name,omitempty"`
	// If present, contains title of the iOS VC in which the Log occured.
	ViewControllerTitle *string `json:"view_controller_title,omitempty"`
	// If present, the Log entry represents a gap in logs reporting.
	// The field marks time at which the gap started.
	GapStart *time.Time `json:"gap_start,omitempty"`
	// Time at which the gap in logs reporting ended.
	GapEnd *time.Time `json:"gap_end,omitempty"`
	// If present, the Log entry contains value for a custom key:value pair.
	// The field contains the key.
	KeyValueKey *string `json:"key_value_key,omitempty"`
	// Value of the key denoted by `KeyValueKey` field.
	KeyValueValue *string `json:"key_value_value,omitempty"`
	// If present, contains information about callback
	// (e.g. Android event such as `OnClick` or iOS selector subscription) execution during which the Log occurred.
	// The field contains the target class name.
	InteractionClass *string `json:"interaction_class,omitempty"`
	// Name of the fired function/method.
	InteractionEventName *string `json:"interaction_event_name,omitempty"`
	// Name of the class which sent the callback.
	InteractionSender *string `json:"interaction_sender,omitempty"`
	// Event type dependent detail(s) of the callback, such as `id`, `title`, `position`, etc.
	InteractionDetail *string `json:"interaction_detail,omitempty"`
	// JS Element XPath
	JSXPath *string `json:"js_xpath,omitempty"`
}
