package hw10programoptimization

import (
	"bytes"
	"testing"
)

var data = `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

var usersTest = [100_000]User{
	{ID: 1, Name: "Howard Mendoza", Username: "0Oliver", Email: "aliquid_qui_ea@Browsedrive.gov", Phone: "6-866-899-36-79", Password: "InAQJvsq", Address: "Blackbird Place 25"},
	{ID: 2, Name: "Howard Mendoz", Username: "0Oliver", Email: "aliquid_qui_ea@Browsedrive.net", Phone: "6-866-899-36-79", Password: "InAQJvsq", Address: "Blackbird Place 25"},
	{ID: 3, Name: "Simon Mendoz", Username: "0Oliver", Email: "qui_ea@Browsedrive.gov", Phone: "6-866-899-36-79", Password: "InAQJvsq", Address: "Blackbird Place 25"},
}

func BenchmarkGetUsers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getUsers(bytes.NewBufferString(data))
	}
}

func BenchmarkCountDomains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		countDomains(usersTest, "gov")
	}
}
