package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var VERSION = "1.0.0"
var BUILD = 1

type Values struct {
	v        *Validator
	d        []string
	def      *string
	name     string
	optional bool
}

func (v *Values) Def(d string) *Values {
	v.def = &d
	return v
}

func (v *Values) Optional() *Values {
	v.optional = true
	return v
}

func (v *Values) Int() int {
	s := v.String()
	if s == "" && v.optional {
		return 0
	}
	ret, err := strconv.Atoi(s)
	if err != nil {
		v.v.Error(v.name, err.Error())
		return 0
	}
	return ret
}

func (v *Values) Int64() int64 {
	s := v.String()
	if s == "" && v.optional {
		return 0
	}
	ret, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		v.v.Error(v.name, err.Error())
		return 0
	}
	return ret
}

func (v *Values) Float() float64 {
	s := v.String()
	if s == "" && v.optional {
		return 0
	}
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		v.v.Error(v.name, err.Error())
		return 0
	}
	return ret
}

func (v *Values) ID() primitive.ObjectID {
	s := v.String()
	if s == "" && v.optional {
		return primitive.NilObjectID
	}
	ret, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		v.v.Error(v.name, err.Error())
	}
	return ret
}

func (v *Values) Bool() bool {
	s := v.String()
	if s == "" && v.optional {
		return false
	}
	ret, err := strconv.ParseBool(s)
	if err != nil {
		v.v.Error(v.name, err.Error())
		return false
	}
	return ret
}

func (v *Values) DateTime() time.Time {
	s := v.String()
	var t time.Time
	if s == "" && v.optional {
		return t
	}
	err := t.UnmarshalText([]byte(s))
	if err != nil {
		v.v.Error(v.name, err.Error())
	}
	return t
}

func (v *Values) String() string {
	if len(v.d) == 0 {
		if v.def == nil {
			if !v.optional {
				v.v.Error(v.name, "Missing Value for "+v.name)
			}
			return ""
		}
		return *v.def
	}
	return v.d[0]
}

func (v *Values) Email() string {
	email := v.String()
	if email == "" && v.optional {
		return ""
	}
	if email == "" {
		v.v.Error(v.name, "Missing Email for "+v.name)
		return ""
	}
	if checkEmail, _ := regexp.MatchString(`^[A-Za-z0-9_.]+@[A-Za-z]{2,}\.[A-Za-z]{2,5}\z`, email); !checkEmail {
		v.v.Error(v.name, "Invalid Email for "+v.name)
		return ""
	}
	return email
}

func (v *Values) Phone() string {
	phone := v.String()
	if phone == "" && v.optional {
		return ""
	}
	if phone == "" {
		v.v.Error(v.name, "Missing Phone for "+v.name)
		return ""
	}
	if checkPhone, _ := regexp.MatchString(`^[0-9]{10}\z`, phone); !checkPhone {
		v.v.Error(v.name, "Invalid Phone for "+v.name)
		return ""
	}
	return phone
}

func (v *Values) IntArray() []int {
	ret := make([]int, len(v.d))
	var err error
	for i := range v.d {
		ret[i], err = strconv.Atoi(v.d[i])
		if err != nil {
			v.v.Error(v.name, err.Error())
			return nil
		}
	}
	return ret
}

func (v *Values) StringArray() []string {
	return v.d
}

func (v *Values) CallEventType() CallEventType {
	ret, err := CallEventTypeFromInt(v.Int())
	if err != nil {
		v.v.Error(v.name, err.Error())
	}
	return ret
}

func (v *Values) CallEventTypeArray() []CallEventType {
	ints := v.IntArray()
	if ints == nil {
		return nil
	}
	var ret []CallEventType
	for _, i := range ints {
		val, err := CallEventTypeFromInt(i)
		if err != nil {
			v.v.Error(v.name, err.Error())
			return nil
		}
		ret = append(ret, val)
	}
	return ret
}

func (v *Values) UserType() UserType {
	ret, err := UserTypeFromInt(v.Int())
	if err != nil {
		v.v.Error(v.name, err.Error())
	}
	return ret
}

func (v *Values) UserTypeArray() []UserType {
	ints := v.IntArray()
	if ints == nil {
		return nil
	}
	var ret []UserType
	for _, i := range ints {
		val, err := UserTypeFromInt(i)
		if err != nil {
			v.v.Error(v.name, err.Error())
			return nil
		}
		ret = append(ret, val)
	}
	return ret
}

func (v *Values) ReqStatus() ReqStatus {
	ret, err := ReqStatusFromInt(v.Int())
	if err != nil {
		v.v.Error(v.name, err.Error())
	}
	return ret
}

func (v *Values) ReqStatusArray() []ReqStatus {
	ints := v.IntArray()
	if ints == nil {
		return nil
	}
	var ret []ReqStatus
	for _, i := range ints {
		val, err := ReqStatusFromInt(i)
		if err != nil {
			v.v.Error(v.name, err.Error())
			return nil
		}
		ret = append(ret, val)
	}
	return ret
}

// -- more-values --
// -- end --

type Validator struct {
	r       *http.Request
	m       map[string]string // Mux Vars
	q       url.Values        // Query Vars
	j       jwt.MapClaims     // JWT Claims
	values  []string
	errors  map[string]string
	_secret string
}

func NewValidator(r *http.Request) *Validator {
	return &Validator{
		r: r,
	}
}

func (v *Validator) Secret(secret string) *Validator {
	v._secret = secret
	return v
}

func (v *Validator) Error(name string, msg string) {
	if v.errors == nil {
		v.errors = make(map[string]string)
	}
	v.errors[name] = msg
}

func (v *Validator) Path(name string) *Values {
	if v.m == nil {
		v.m = mux.Vars(v.r)
	}
	var d []string
	if data, ok := v.m[name]; ok {
		d = append(d, data)
	}
	return &Values{
		name: name,
		v:    v,
		d:    d,
	}
}

func (v *Validator) nilValues(name string) *Values {
	return &Values{
		name: name,
		v:    v,
		d:    nil,
	}
}

func (v *Validator) jwtSugarFn(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
	}

	return []byte(v._secret), nil
}

func (v *Validator) Token(name string) *Values {
	if v.j == nil {
		tokenStr := v.r.Header.Get("jwt")
		if tokenStr == "" {
			tokenStr = v.Query("__jwt").Optional().String()
			if tokenStr == "" {
				v.Error(name, "Missing token")
				return v.nilValues(name)
			}
		}
		if tokenStr != "" && tokenStr != "null" && tokenStr != "nil" {
			token, err := jwt.Parse(tokenStr, v.jwtSugarFn)
			if err != nil {
				v.Error(name, err.Error())
				return v.nilValues(name)
			}

			var ok bool
			v.j, ok = token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				v.Error(name, "Could not parse claims or token not valid")
				return v.nilValues(name)
			}
		}
	}

	expiry, hasExpiry := v.j["expiry"]
	if hasExpiry {
		unixExpiry, err := strconv.Atoi(expiry.(string))
		if err == nil {
			t := time.Unix(int64(unixExpiry), 0)
			if t.Before(time.Now()) {
				v.Error(name, "Token Expired")
				return v.nilValues(name)
			}
		}
	}

	var d []string
	value, ok := v.j[name]
	if !ok {
		return v.nilValues(name)
	}
	d = append(d, value.(string))
	return &Values{
		name: name,
		v:    v,
		d:    d,
	}
}

func (v *Validator) Query(name string) *Values {
	if v.q == nil {
		v.q = v.r.URL.Query()
	}
	return &Values{
		name: name,
		v:    v,
		d:    v.q[name],
	}
}

func (v *Validator) Form(name string) *Values {
	v.r.ParseForm()
	return &Values{
		name: name,
		v:    v,
		d:    v.r.Form[name],
	}
}

func (v *Validator) HasForm(name string) bool {
	v.r.ParseForm()
	return len(v.r.Form[name]) > 0
}

func (v *Validator) HasQuery(name string) bool {
	if v.q == nil {
		v.q = v.r.URL.Query()
	}
	return len(v.q[name]) > 0
}

func (v *Validator) Has(name string) bool {
	return v.HasForm(name) || v.HasQuery(name)
}

// File reads up the file from underlying request and returns it.
func (v *Values) File() []byte {
	f, _, err := v.v.r.FormFile(v.name)
	if err != nil && !v.optional {
		v.v.Error(v.name, err.Error())
		return nil
	}
	defer f.Close()

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(f)
	if err != nil && !v.optional {
		v.v.Error(v.name, err.Error())
	}
	return buffer.Bytes()
}

func (v *Validator) Header(name string) *Values {
	return &Values{
		name: name,
		v:    v,
		d:    v.r.Header[name],
	}
}

func (v *Validator) Cookie(name string) *Values {
	var d []string
	if data, err := v.r.Cookie(name); err == nil {
		d = append(d, data.Value)
	}
	return &Values{
		name: name,
		v:    v,
		d:    d,
	}
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

func (v *Validator) Write(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	b, _ := json.Marshal(v.errors)
	w.Write(b)
}

func (v *Validator) Errors() map[string]string {
	return v.errors
}

func (v *Validator) GetError() string {
	var ret string
	for k, v := range v.errors {
		ret = ret + fmt.Sprintf("%s=%s,", k, v)
	}
	return ret
}
