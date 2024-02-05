// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package myjsons

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons(in *jlexer.Lexer, out *StoreItem) {
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
		case "uuid":
			out.ID = int(in.Int())
		case "short_url":
			out.ShortURL = string(in.String())
		case "original_url":
			out.OriginalURL = string(in.String())
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
func easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons(out *jwriter.Writer, in StoreItem) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"uuid\":"
		out.RawString(prefix[1:])
		out.Int(int(in.ID))
	}
	{
		const prefix string = ",\"short_url\":"
		out.RawString(prefix)
		out.String(string(in.ShortURL))
	}
	{
		const prefix string = ",\"original_url\":"
		out.RawString(prefix)
		out.String(string(in.OriginalURL))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v StoreItem) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v StoreItem) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *StoreItem) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *StoreItem) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons(l, v)
}
func easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons1(in *jlexer.Lexer, out *Response) {
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
		case "result":
			out.Result = string(in.String())
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
func easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons1(out *jwriter.Writer, in Response) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"result\":"
		out.RawString(prefix[1:])
		out.String(string(in.Result))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Response) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Response) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Response) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Response) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons1(l, v)
}
func easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons2(in *jlexer.Lexer, out *Request) {
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
		case "url":
			out.URL = string(in.String())
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
func easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons2(out *jwriter.Writer, in Request) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"url\":"
		out.RawString(prefix[1:])
		out.String(string(in.URL))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Request) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Request) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonEd95e889EncodeGithubComDmad1989UrlcutInternalMyjsons2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Request) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Request) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonEd95e889DecodeGithubComDmad1989UrlcutInternalMyjsons2(l, v)
}
