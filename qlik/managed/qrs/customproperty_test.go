package qrs

import (
	"github.com/Click-CI/common/util"
	"reflect"
	"testing"
)

func TestClient_GetCustomPropertyList(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)

	cps, res := client.GetCustomPropertyList()
	if res != nil {
		t.Errorf("failed: %s", res.Error())
		return
	}

	logger.Info().Msg(string(util.Jsonify(&cps)))
}

func TestCPMap_diff(t *testing.T) {
	v1 := &CustomPropertyValue{Value: "v1"}
	v2 := &CustomPropertyValue{Value: "v2"}
	v3 := &CustomPropertyValue{Value: "v3"}
	v4 := &CustomPropertyValue{Value: "v4"}
	m1 := CPValueMap{"k1": v1, "k2": v2}
	m2 := CPValueMap{"k1": v1}
	m3 := CPValueMap{"k1": v1, "k2": v2, "k3": v3}
	m4 := m1
	m5 := CPValueMap{"k1": v1, "k3": v2, "k4": v4}
	type args struct {
		right CPValueMap
	}
	tests := []struct {
		name string
		left CPValueMap
		args args
		want []MapDiffValue
	}{
		{name: "equal", left: m1, args: args{right: m4}, want: []MapDiffValue{}},
		{name: "left", left: m1, args: args{right: m2}, want: []MapDiffValue{MapDiffValue{Type: -1, Key: "k2", Left: v2, Right: nil}}},
		{name: "right", left: m1, args: args{right: m3}, want: []MapDiffValue{MapDiffValue{Type: 1, Key: "k3", Left: nil, Right: v3}}},
		{
			name: "mix",
			left: m3, args: args{right: m5},
			want: []MapDiffValue{
				MapDiffValue{Type: -1, Key: "k2", Left: v2, Right: nil},
				MapDiffValue{Type: 0, Key: "k3", Left: v3, Right: v2},
				MapDiffValue{Type: 1, Key: "k4", Left: nil, Right: v4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.left.Diff(tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_AddAppCustomProperty(t *testing.T) {
	client, _, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)

	app, res := client.GetApp(happinessAppID) //Happiness
	if res != nil {
		t.Errorf("failed: %s", res.Error())
	}

	res = client.AddAppCustomProperty(app, "haha", "hah1")
	if res != nil {
		t.Errorf("AddAppCustomProperty failed: %s", res.Error())
	}
}
