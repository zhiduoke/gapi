package pbjson

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/pbjson"
	"github.com/zhiduoke/gapi/proto/pdparser"
	"github.com/zhiduoke/gapi/tests/msgs"
)

const benchJSON = `{
  "msgs": [
    {
      "id": "5e2db5de0682b0751526db6f",
      "index": 0,
      "guid": "dfcfd9a6-fb4a-4058-ac1d-0c0bcb8413a4",
      "isActive": true,
      "balance": "$2,342.00",
      "picture": "http://placehold.it/32x32",
      "age": 37,
      "eyeColor": "brown",
      "name": "Arlene Carney",
      "gender": "female",
      "company": "POOCHIES",
      "email": "arlenecarney@poochies.com",
      "phone": "+1 (923) 462-2911",
      "address": "400 Strickland Avenue, Idamay, Northern Mariana Islands, 5954",
      "about": "Magna aliquip ut nulla amet minim proident proident. Voluptate mollit pariatur amet occaecat consectetur ad deserunt incididunt laborum dolore. Commodo fugiat ipsum pariatur officia occaecat adipisicing consequat exercitation ipsum. Ullamco tempor qui eiusmod cupidatat sint aute sint esse fugiat deserunt. Occaecat deserunt cillum non ullamco culpa voluptate. Cupidatat aute exercitation ex pariatur culpa anim laborum. Veniam dolor fugiat pariatur aliqua aute aute velit incididunt sunt eiusmod non ea reprehenderit ullamco.\r\n",
      "registered": "2014-09-08T03:32:41 -08:00",
      "latitude": 87.14156,
      "longitude": 70.879769,
      "tags": [
        "ex",
        "consequat",
        "eiusmod",
        "id",
        "duis",
        "id",
        "duis"
      ],
      "friends": [
        {
          "id": 0,
          "name": "Slater Rosales"
        },
        {
          "id": 1,
          "name": "Angelique Christian"
        },
        {
          "id": 2,
          "name": "Luna Aguilar"
        }
      ],
      "greeting": "Hello, Arlene Carney! You have 5 unread messages.",
      "favoriteFruit": "apple"
    },
    {
      "id": "5e2db5de73de61659ad4758a",
      "index": 1,
      "guid": "33b56ade-4c4f-4508-974e-4870dd49ff33",
      "isActive": false,
      "balance": "$3,490.16",
      "picture": "http://placehold.it/32x32",
      "age": 28,
      "eyeColor": "blue",
      "name": "Miller Munoz",
      "gender": "male",
      "company": "BRAINQUIL",
      "email": "millermunoz@brainquil.com",
      "phone": "+1 (873) 475-3519",
      "address": "747 Commercial Street, Ola, Washington, 6285",
      "about": "Anim in anim consequat Lorem tempor excepteur do commodo duis cupidatat minim. Commodo minim velit excepteur enim ea consectetur commodo officia. Duis elit amet sit anim esse. Ullamco ad cupidatat et cillum culpa et ipsum tempor magna sint quis deserunt dolore nostrud.\r\n",
      "registered": "2014-06-07T12:58:00 -08:00",
      "latitude": 4.329646,
      "longitude": -85.269235,
      "tags": [
        "consectetur",
        "mollit",
        "elit",
        "duis",
        "nulla",
        "enim",
        "dolor"
      ],
      "friends": [
        {
          "id": 0,
          "name": "Castaneda Terry"
        },
        {
          "id": 1,
          "name": "Lisa Cunningham"
        },
        {
          "id": 2,
          "name": "Madden Delaney"
        }
      ],
      "greeting": "Hello, Miller Munoz! You have 3 unread messages.",
      "favoriteFruit": "strawberry"
    },
    {
      "id": "5e2db5de4a0a200b33fb80f0",
      "index": 2,
      "guid": "b822bc28-222f-4b03-b8b7-1cf89452e910",
      "isActive": false,
      "balance": "$2,660.05",
      "picture": "http://placehold.it/32x32",
      "age": 29,
      "eyeColor": "brown",
      "name": "Delacruz Barnes",
      "gender": "male",
      "company": "INSURON",
      "email": "delacruzbarnes@insuron.com",
      "phone": "+1 (897) 459-2474",
      "address": "540 Hancock Street, Sunbury, Virgin Islands, 931",
      "about": "Esse qui sit deserunt esse in velit. Consectetur nostrud adipisicing duis non ut. Tempor pariatur cillum reprehenderit dolor ea commodo laborum et culpa labore ullamco tempor. Ullamco culpa reprehenderit tempor ea nulla. Pariatur ea fugiat excepteur amet quis excepteur laboris. Dolore qui ipsum in qui nulla. Culpa amet non Lorem in cupidatat exercitation eiusmod nulla quis ipsum.\r\n",
      "registered": "2014-01-21T07:52:14 -08:00",
      "latitude": -87.652499,
      "longitude": 89.279269,
      "tags": [
        "consequat",
        "anim",
        "commodo",
        "non",
        "ullamco",
        "fugiat",
        "mollit"
      ],
      "friends": [
        {
          "id": 0,
          "name": "Morales Mcclain"
        },
        {
          "id": 1,
          "name": "Lynda Osborne"
        },
        {
          "id": 2,
          "name": "Terri Frost"
        }
      ],
      "greeting": "Hello, Delacruz Barnes! You have 2 unread messages.",
      "favoriteFruit": "banana"
    },
    {
      "id": "5e2db5de572a3938461228f7",
      "index": 3,
      "guid": "e089b69b-d818-4366-9f0b-c46fc1e313e5",
      "isActive": false,
      "balance": "$1,565.53",
      "picture": "http://placehold.it/32x32",
      "age": 22,
      "eyeColor": "brown",
      "name": "Tran Holt",
      "gender": "male",
      "company": "LUXURIA",
      "email": "tranholt@luxuria.com",
      "phone": "+1 (877) 552-3299",
      "address": "797 Franklin Avenue, Turpin, Nebraska, 6457",
      "about": "Eiusmod eiusmod irure exercitation mollit ea est incididunt aliquip laboris ad. Proident reprehenderit reprehenderit et ea mollit. Nulla laborum ex amet cillum excepteur laboris. Excepteur exercitation duis anim pariatur elit aliqua reprehenderit sit fugiat nulla labore ex.\r\n",
      "registered": "2015-05-12T04:35:09 -08:00",
      "latitude": 68.777412,
      "longitude": 13.365268,
      "tags": [
        "velit",
        "Lorem",
        "velit",
        "ullamco",
        "laborum",
        "ad",
        "veniam"
      ],
      "friends": [
        {
          "id": 0,
          "name": "Morgan Price"
        },
        {
          "id": 1,
          "name": "Louisa Livingston"
        },
        {
          "id": 2,
          "name": "Bridget Vang"
        }
      ],
      "greeting": "Hello, Tran Holt! You have 10 unread messages.",
      "favoriteFruit": "strawberry"
    },
    {
      "id": "5e2db5de3748e3be2a32d69f",
      "index": 4,
      "guid": "98788a86-9511-4814-b684-242dbabe0be4",
      "isActive": true,
      "balance": "$1,907.74",
      "picture": "http://placehold.it/32x32",
      "age": 39,
      "eyeColor": "green",
      "name": "Janna Murphy",
      "gender": "female",
      "company": "GORGANIC",
      "email": "jannamurphy@gorganic.com",
      "phone": "+1 (816) 450-2027",
      "address": "202 Robert Street, Falconaire, Montana, 3007",
      "about": "Ex in magna non duis cillum enim exercitation cillum commodo aliqua exercitation aliquip cillum adipisicing. Qui non minim officia qui Lorem duis ex consequat commodo elit ipsum commodo eu. Ea aliquip pariatur Lorem magna consectetur deserunt.\r\n",
      "registered": "2019-04-22T11:22:28 -08:00",
      "latitude": -80.510177,
      "longitude": 65.630181,
      "tags": [
        "tempor",
        "elit",
        "in",
        "incididunt",
        "sunt",
        "velit",
        "Lorem"
      ],
      "friends": [
        {
          "id": 0,
          "name": "Roy Gates"
        },
        {
          "id": 1,
          "name": "Caitlin Clements"
        },
        {
          "id": 2,
          "name": "Elsie Cain"
        }
      ],
      "greeting": "Hello, Janna Murphy! You have 9 unread messages.",
      "favoriteFruit": "apple"
    }
  ]
}`

var (
	benchMsg    *metadata.Message
	benchPBData []byte
)

func init() {
	pddata, err := ioutil.ReadFile("../msgs/bench.pd")
	if err != nil {
		log.Fatal(err)
	}
	md, err := pdparser.ParseSet(pddata)
	if err != nil {
		log.Fatal(err)
	}
	benchMsg = md.Routes[0].Call.Out
	var reply msgs.BenchReply
	err = json.Unmarshal([]byte(benchJSON), &reply)
	if err != nil {
		log.Fatal(err)
	}
	benchPBData, _ = proto.Marshal(&reply)
}

func encode() ([]byte, error) {
	e := pbjson.NewEncoder(make([]byte, 0, len(benchPBData)*2))
	e.EncodeMessage(benchMsg, benchPBData)
	if e.Error() != nil {
		return nil, e.Error()
	}
	return e.Bytes(), nil
}

func encodeFast() ([]byte, error) {
	e := pbjson.NewEncoder(make([]byte, 0, len(benchPBData)*2))
	e.EncodeMessageFast(benchMsg, benchPBData)
	if e.Error() != nil {
		return nil, e.Error()
	}
	return e.Bytes(), nil
}

func encodeStd() ([]byte, error) {
	var reply msgs.BenchReply
	err := proto.Unmarshal(benchPBData, &reply)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&reply)
}

func TestPBToJson(t *testing.T) {
	out, err := encode()
	if err != nil {
		t.Fatal(err)
	}
	var reply msgs.BenchReply
	err = json.Unmarshal(out, &reply)
	benchPBData2, _ := proto.Marshal(&reply)
	if !bytes.Equal(benchPBData, benchPBData2) {
		t.Log(hex.EncodeToString(benchPBData))
		t.Log(hex.EncodeToString(benchPBData2))
		t.Fatal("not equal")
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encode()
	}
}

func BenchmarkEncodeFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encodeFast()
	}
}

func BenchmarkEncodeStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encodeStd()
	}
}
