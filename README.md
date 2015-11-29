API written in Go for www.pictriev.com
--------------------------------------

# Note
이 API는 공식적인 API가 아닙니다. 단순히 서버에 원격 요청을 보내는 것이므로
인터넷 사용을 필요로 합니다.

# Example
## Parse image
```go
...
parsed, err := FindFaceImage(src)
if err != nil {
	log.Fatal(err)
}

if parsed.NFaces == 0 {
	// No face in the image
	...
}

// At least 1 faces in the image
...
```

`FindFace*` 함수를 사용해서 이미지를 분석할 수 있습니다.
`FindFaceImage` 함수는 이미지를 서버에 전송함으로써 이미지를 분석합니다. 반면에
`FindFaceImageURL` 함수는 이미지 주소를 보냄으로써 이미지를 분석합니다. 이미지 접근에
특별한 권한이 요구된다면 이 함수는 실패할 수 있습니다.

## Whois
```go
// inquiry all faces in the image sent. 
for i := 0; i < parsed.NFaces; i++ {
	whois, err := Whois(result.ImageID, i, Ko)
	if err != nil {
		log.Fatal(err)
	}

	...
}
```

`Whois` 함수로 분석 결과를 받아올 수 있습니다. 다음은 `WhoisResult` 각각의 필드에 대한
설명입니다.

	Age     float64   `json:"age"`
	AgeDist []float64 `json:"agedist"`
	// Attrs 는 높은 확률 순서로 정렬되어 있다.
	Attrs  []Attr
	Gender Gender
	Lang   string

- **Age**: 평균 나이를 나타냅니다.
- **AgeDist**: 각 나이대별 확률 분포를 나타냅니다.
  순서대로 [0, 10), [10, 20), ... 나이 그룹을 의미합니다.
- **Attrs**: 누구와 닮았는지에 대한 정보를 포함하고 있습니다.
- **Gender**: 확률적으로 성별을 의미합니다.
- **Lang**: 사용된 언어를 나타냅니다.

## Example
```go
func main() {
	...
	result, err := FindFaceImage(src)
	handle_err(err)

	if result.NFaces == 0 {
		fmt.Println("얼굴이 없음")
		return
	}

	fmt.Printf("수행 시간 %.4f초\n", result.PTime)
	fmt.Printf("이미지 크기 %dx%d\n", result.Width, result.Height)
	fmt.Printf("%d개의 얼굴 검색됨\n", result.NFaces)

	for i := 0; i < result.NFaces; i++ {
		whois, err := Whois(result.ImageID, i, Ko)
		handle_err(err)

		fmt.Println()
		fmt.Printf("%d번째 얼굴 결과\n", i+1)
		fmt.Printf("평균 나이 %.5f\n", whois.Age)
		if whois.Gender.Man() > 0.5 {
			fmt.Printf("남자 (%.5f)\n", whois.Gender.Man())
		} else {
			fmt.Printf("여자 (%.5f)\n", whois.Gender.Woman())
		}
		for i, attr := range whois.Attrs {
			fmt.Printf("  %2d %s %.4f\n", i+1, attr.Name, attr.P)
		}
	}
}
```
