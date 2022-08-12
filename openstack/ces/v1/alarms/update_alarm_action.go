package alarms

import golangsdk "github.com/opentelekomcloud/gophertelekomcloud"

type ModifyAlarmActionRequest struct {
	// Specifies whether the alarm rule is enabled.
	AlarmEnabled bool `json:"alarm_enabled"`
}

func UpdateAlarmAction(client *golangsdk.ServiceClient, id string, req ModifyAlarmActionRequest) (err error) {
	reqBody, err := golangsdk.BuildRequestBody(req, "")
	if err != nil {
		return
	}

	// PUT /V1.0/{project_id}/alarms/{alarm_id}/action
	_, err = client.Put(client.ServiceURL("alarms", id, "action"), reqBody, nil, &golangsdk.RequestOpts{
		OkCodes: []int{204},
	})
	return
}
