// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package internal

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	time "time"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal(in *jlexer.Lexer, out *brokerMessage) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "company_name":
			out.CompanyName = string(in.String())
		case "client_id":
			out.ClientId = string(in.String())
		case "message":
			out.Message = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal(out *jwriter.Writer, in brokerMessage) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix[1:])
		out.String(string(in.CompanyName))
	}
	{
		const prefix string = ",\"client_id\":"
		out.RawString(prefix)
		out.String(string(in.ClientId))
	}
	{
		const prefix string = ",\"message\":"
		out.RawString(prefix)
		out.String(string(in.Message))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v brokerMessage) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v brokerMessage) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *brokerMessage) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *brokerMessage) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal1(in *jlexer.Lexer, out *CompanyToken) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "token":
			out.Token = string(in.String())
		case "server_name":
			out.ServerName = string(in.String())
		case "duration_hour":
			out.Duration = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal1(out *jwriter.Writer, in CompanyToken) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"token\":"
		out.RawString(prefix[1:])
		out.String(string(in.Token))
	}
	{
		const prefix string = ",\"server_name\":"
		out.RawString(prefix)
		out.String(string(in.ServerName))
	}
	{
		const prefix string = ",\"duration_hour\":"
		out.RawString(prefix)
		out.Int(int(in.Duration))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CompanyToken) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v CompanyToken) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CompanyToken) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *CompanyToken) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal1(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal2(in *jlexer.Lexer, out *CompanyStatsWrapper) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "stats":
			if in.IsNull() {
				in.Skip()
				out.Stats = nil
			} else {
				in.Delim('[')
				if out.Stats == nil {
					if !in.IsDelim(']') {
						out.Stats = make([]CompanyStats, 0, 0)
					} else {
						out.Stats = []CompanyStats{}
					}
				} else {
					out.Stats = (out.Stats)[:0]
				}
				for !in.IsDelim(']') {
					var v1 CompanyStats
					(v1).UnmarshalEasyJSON(in)
					out.Stats = append(out.Stats, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal2(out *jwriter.Writer, in CompanyStatsWrapper) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"stats\":"
		out.RawString(prefix[1:])
		if in.Stats == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Stats {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CompanyStatsWrapper) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v CompanyStatsWrapper) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CompanyStatsWrapper) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *CompanyStatsWrapper) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal2(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal3(in *jlexer.Lexer, out *CompanyStats) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "company_name":
			out.Name = string(in.String())
		case "online_users":
			out.OnlineUsers = uint(in.Uint())
		case "max_users":
			out.MaxUsers = uint(in.Uint())
		case "active_users":
			if in.IsNull() {
				in.Skip()
				out.Users = nil
			} else {
				in.Delim('[')
				if out.Users == nil {
					if !in.IsDelim(']') {
						out.Users = make([]ClientStat, 0, 4)
					} else {
						out.Users = []ClientStat{}
					}
				} else {
					out.Users = (out.Users)[:0]
				}
				for !in.IsDelim(']') {
					var v4 ClientStat
					(v4).UnmarshalEasyJSON(in)
					out.Users = append(out.Users, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "ttl":
			out.TTL = time.Duration(in.Int64())
		case "time":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Time).UnmarshalJSON(data))
			}
		case "stoped_time":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Stopped).UnmarshalJSON(data))
			}
		case "expired":
			out.Expired = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal3(out *jwriter.Writer, in CompanyStats) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix[1:])
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"online_users\":"
		out.RawString(prefix)
		out.Uint(uint(in.OnlineUsers))
	}
	{
		const prefix string = ",\"max_users\":"
		out.RawString(prefix)
		out.Uint(uint(in.MaxUsers))
	}
	{
		const prefix string = ",\"active_users\":"
		out.RawString(prefix)
		if in.Users == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Users {
				if v5 > 0 {
					out.RawByte(',')
				}
				(v6).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"ttl\":"
		out.RawString(prefix)
		out.Int64(int64(in.TTL))
	}
	{
		const prefix string = ",\"time\":"
		out.RawString(prefix)
		out.Raw((in.Time).MarshalJSON())
	}
	{
		const prefix string = ",\"stoped_time\":"
		out.RawString(prefix)
		out.Raw((in.Stopped).MarshalJSON())
	}
	{
		const prefix string = ",\"expired\":"
		out.RawString(prefix)
		out.Bool(bool(in.Expired))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CompanyStats) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v CompanyStats) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CompanyStats) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *CompanyStats) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal3(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal4(in *jlexer.Lexer, out *ClientStat) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "user_id":
			out.UserId = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal4(out *jwriter.Writer, in ClientStat) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"user_id\":"
		out.RawString(prefix[1:])
		out.String(string(in.UserId))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ClientStat) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ClientStat) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ClientStat) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ClientStat) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal4(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal5(in *jlexer.Lexer, out *AddCompanyResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "company_token":
			(out.Token).UnmarshalEasyJSON(in)
		case "company_name":
			out.CompanyName = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal5(out *jwriter.Writer, in AddCompanyResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"company_token\":"
		out.RawString(prefix[1:])
		(in.Token).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix)
		out.String(string(in.CompanyName))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v AddCompanyResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v AddCompanyResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *AddCompanyResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *AddCompanyResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal5(l, v)
}
func easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal6(in *jlexer.Lexer, out *AddCompanyMessage) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "company_name":
			out.CompanyName = string(in.String())
		case "max_users":
			out.MaxUsers = uint(in.Uint())
		case "duration_hour":
			out.Duration = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal6(out *jwriter.Writer, in AddCompanyMessage) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix[1:])
		out.String(string(in.CompanyName))
	}
	{
		const prefix string = ",\"max_users\":"
		out.RawString(prefix)
		out.Uint(uint(in.MaxUsers))
	}
	{
		const prefix string = ",\"duration_hour\":"
		out.RawString(prefix)
		out.Int(int(in.Duration))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v AddCompanyMessage) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal6(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v AddCompanyMessage) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGithubComDmitryDmsAvalancheInternal6(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *AddCompanyMessage) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal6(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *AddCompanyMessage) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGithubComDmitryDmsAvalancheInternal6(l, v)
}
