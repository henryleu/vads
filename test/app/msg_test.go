package app

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func genTestRegistration() (*Registration, []byte) {
	reg := &Registration{TeamID: 1, TeamName: "daolaji"}
	bytes, _ := json.Marshal(reg.Message())
	head := fmt.Sprintf("%05d", len(bytes))
	wiredBytes := append([]byte(head), bytes...)
	fmt.Println(reg)
	fmt.Println(string(wiredBytes))
	return reg, wiredBytes
}

func genTestGameOver() (*GameOver, []byte) {
	gameOver := &GameOver{}
	bytes, _ := json.Marshal(gameOver.Message())
	head := fmt.Sprintf("%05d", len(bytes))
	wiredBytes := append([]byte(head), bytes...)
	fmt.Println(gameOver)
	fmt.Println(string(wiredBytes))
	return gameOver, wiredBytes
}

func TestParseMessageOnWire(t *testing.T) {
	reg, regBytes := genTestRegistration()
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Message
		wantErr bool
	}{
		{
			name: "succeed to parse registration message",
			args: args{regBytes},
			want: &Message{
				Name:    RegistrationName,
				Payload: reg,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessageOnWire(tt.args.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessageOnWire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMessageOnWire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_BytesOnWire(t *testing.T) {
	reg, regBytes := genTestRegistration()
	type fields struct {
		Name    string
		Payload Payload
		Raw     bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "BytesOnWire - succeed to parse a registration message from bytes on wire",
			fields: fields{
				Name:    RegistrationName,
				Payload: reg,
				Raw:     false,
			},
			want:    regBytes,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Name:    tt.fields.Name,
				Payload: tt.fields.Payload,
				Raw:     tt.fields.Raw,
			}
			got, err := m.BytesOnWire()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.BytesOnWire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Message.BytesOnWire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_Registration(t *testing.T) {
	reg, regBytes := genTestRegistration()
	regMsg, _ := ParseMessageOnWire(regBytes)
	_, overBytes := genTestGameOver()
	overMsg, _ := ParseMessageOnWire(overBytes)
	type fields struct {
		Name    string
		Payload Payload
		Raw     bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Registration
		wantErr bool
	}{
		{
			name: "Registration - succeed to return a registration",
			fields: fields{
				Name:    regMsg.Name,
				Payload: regMsg.Payload,
				Raw:     regMsg.Raw,
			},
			want:    reg,
			wantErr: false,
		},
		{
			name: "Registration - fail to return a registration",
			fields: fields{
				Name:    overMsg.Name,
				Payload: overMsg.Payload,
				Raw:     overMsg.Raw,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Name:    tt.fields.Name,
				Payload: tt.fields.Payload,
				Raw:     tt.fields.Raw,
			}
			got, err := m.Registration()
			// fmt.Println(got)
			// fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Registration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Message.Registration() = %v, want %v", got, tt.want)
			}
		})
	}
}
