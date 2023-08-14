package qrs

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/rac"
	"net/http"
	"strings"
	"time"
)

type Subscription struct {
	TypeName     *string `json:"type_name,omitempty" yaml:"type_name,omitempty"`
	ObjectId     *string `json:"object_id,omitempty" yaml:"object_id,omitempty"`
	Filter       *string `json:"filter,omitempty" yaml:"filter,omitempty"`
	Condition    *string `json:"condition,omitempty" yaml:"condition,omitempty"`
	ChangeType   *string `json:"change_type,omitempty" yaml:"change_type,omitempty"`
	PropertyName *string `json:"property_name,omitempty" yaml:"property_name,omitempty"`
	CallbackURL  string  `json:"callback_url,omitempty" yaml:"callback_url,omitempty"`
}

func (s Subscription) GetParams() map[string]string {
	ret := make(map[string]string)
	if s.TypeName != nil {
		ret["name"] = *s.TypeName
	}
	if s.ObjectId != nil {
		ret["id"] = *s.ObjectId
	}
	if s.Filter != nil {
		ret["filter"] = *s.Filter
	}
	if s.Condition != nil {
		ret["condition"] = *s.Condition
	}
	if s.ChangeType != nil {
		ret["changetype"] = *s.ChangeType
	}
	if s.PropertyName != nil {
		ret["propertyname"] = *s.PropertyName
	}
	return ret
}

func (s Subscription) Valid() *util.Result {
	if s.CallbackURL == "" {
		return util.Errorf("callback url is empty")
	}
	return nil
}

type SubscriptionResponse struct {
	Value string
}

type ChangeType int

func (t ChangeType) String() string {
	switch int(t) {
	case 1:
		return "Add"
	case 2:
		return "Update"
	case 3:
		return "Delete"
	default:
		return "Undefined"
	}
}

func (t *ChangeType) MarshalJSON() ([]byte, error) {
	v := t.String()
	return json.Marshal(&v)
}

type ChangeEvent struct {
	ChangeType          ChangeType `json:"changeType,omitempty"`
	ObjectType          string     `json:"objectType,omitempty"`
	ObjectID            string     `json:"objectID,omitempty"`
	ChangedProperties   []string   `json:"changedProperties,omitempty"`
	EngineID            string     `json:"engineID,omitempty"`
	EngineType          string     `json:"engineType,omitempty"`
	OriginatorNodeID    string     `json:"originatorNodeID,omitempty"`
	OriginatorHostName  string     `json:"originatorHostName,omitempty"`
	OriginatorContextID string     `json:"originatorContextID,omitempty"`
	CreatedDate         time.Time  `json:"createdDate,omitempty"`
	ModifiedDate        time.Time  `json:"modifiedDate,omitempty"`
	SchemaPath          string     `json:"schemaPath,omitempty"`
}

type ChangeEvents []ChangeEvent

func (c *Client) Subscribe(sub Subscription) (string, *util.Result) {
	cbUrl := strings.Trim(sub.CallbackURL, `"`)
	if cbUrl == "" {
		return "", util.MsgError("Check", "no callback url")
	}
	cbUrl = fmt.Sprintf(`"%s"`, cbUrl)
	body := strings.NewReader(cbUrl)
	_, resp, res := c.client.Do(http.MethodPost, "/notification", body, rac.WithParams(sub.GetParams()))
	if res != nil {
		return "", res.With("Do")
	}

	var subId SubscriptionResponse
	err := json.Unmarshal(resp, &subId)
	if err != nil {
		return "", util.Error("ParseId", err)
	}

	return subId.Value, nil
}

type NotiSubscriber struct {
	subs                []*Subscription
	StopNotiSubscriber  chan bool
	SubscriptionHandles []string
	client              *Client
	Logger              *zerolog.Logger
}

func NewNotiSubscriber(cfg Config, subs []*Subscription) (*NotiSubscriber, *util.Result) {
	ret := &NotiSubscriber{
		StopNotiSubscriber:  make(chan bool),
		SubscriptionHandles: make([]string, 0),
		client:              nil,
		Logger:              loggers.NullLogger,
	}

	var res *util.Result
	ret.client, res = NewClient(cfg)
	if res != nil {
		return nil, res.With("NewQrsClient")
	}

	for i, s := range subs {
		if res = s.Valid(); res != nil {
			return nil, res.With(fmt.Sprintf("Sub[%d] is not valid", i))
		}
	}

	return ret, nil
}

func (a *NotiSubscriber) Subscribe() *util.Result {
	a.Logger.Debug().Msgf("start to subscribe to %d subscriptions", len(a.subs))
	a.SubscriptionHandles = make([]string, 0)
	for i, sub := range a.subs {
		subId, res := a.client.Subscribe(*sub)
		if res != nil {
			a.Logger.Err(res).Msgf("subsciption[%d] failed", i)
			return res.With(fmt.Sprintf("Subscribe[%d]", i))
		}
		a.Logger.Debug().Msgf("subscription[%d] handle: %s", i, subId)
		a.SubscriptionHandles = append(a.SubscriptionHandles, subId)
	}
	return nil
}

func (a *NotiSubscriber) StartHealthCheckThread(timerSec int) {
	a.Logger.Debug().Msg("start Notification Auditor")
	ticker := time.NewTicker(time.Duration(timerSec) * time.Second)

	QrsIsDown := false
	for {
		select {
		case _, ok := <-a.StopNotiSubscriber:
			if !ok {
				a.Logger.Warn().Msgf("Notification Auditor stropped in invalid status")
			}
			a.Logger.Info().Msg("Notification Auditor stopped")
			return
		case _ = <-ticker.C:
			a.Logger.Debug().Msgf("QRS is down:? %v, checking health again ...", QrsIsDown)
			_, res := a.client.About()
			if res != nil {
				a.Logger.Error().Msgf("QRS about failed %s ", res.Error())
				QrsIsDown = true
			}

			if QrsIsDown {
				a.Logger.Warn().Msgf("QRS is down, try to re-subscribe ...")
				res = a.Subscribe()
				if res != nil {
					a.Logger.Err(res).Msg("re-subscribe failed, will try again later")
				} else {
					QrsIsDown = false
					a.Logger.Info().Msg("re-subscribe succeeded")
				}
			}
		}
	}
}
