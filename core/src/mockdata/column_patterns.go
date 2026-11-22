package mockdata
import (
	"regexp"
	"strings"
	"github.com/brianvoe/gofakeit/v7"
)
type ColumnPattern struct {
	Pattern  *regexp.Regexp
	Name     string
	Generate func(f *gofakeit.Faker, maxLen int) any
}
var columnPatterns = []ColumnPattern{
	{regexp.MustCompile(`(?i)(^|_)e[-_]?mail(_|$)`), "email", genEmail},
	{regexp.MustCompile(`(?i)(^|_)(user[-_]?name|uname|login)(_|$)`), "username", genUsername},
	{regexp.MustCompile(`(?i)(^|_)(first[-_]?name|fname|given[-_]?name)(_|$)`), "first_name", genFirstName},
	{regexp.MustCompile(`(?i)(^|_)(last[-_]?name|lname|surname|family[-_]?name)(_|$)`), "last_name", genLastName},
	{regexp.MustCompile(`(?i)^name$|(^|_)(full[-_]?name|display[-_]?name)(_|$)`), "full_name", genFullName},
	{regexp.MustCompile(`(?i)(^|_)(phone|mobile|cell|telephone|tel)(_|$)`), "phone", genPhone},
	{regexp.MustCompile(`(?i)(^|_)(ip|ip[-_]?addr(ess)?)(_|$)`), "ip", genIP},
	{regexp.MustCompile(`(?i)^(url|website|link|homepage)$`), "url", genURL},
	{regexp.MustCompile(`(?i)(^|_)(street[-_]?address|address[-_]?line|address|street)(_|$)`), "address", genStreet},
	{regexp.MustCompile(`(?i)^city$`), "city", genCity},
	{regexp.MustCompile(`(?i)^(state|province|region)$`), "state", genState},
	{regexp.MustCompile(`(?i)^country$`), "country", genCountry},
	{regexp.MustCompile(`(?i)(^|_)(zip|postal|postcode)(_|$)`), "zip", genZip},
	{regexp.MustCompile(`(?i)(^|_)(company|organization|org)(_|$)`), "company", genCompany},
	{regexp.MustCompile(`(?i)(^|_)(job[-_]?title|title|position|role)(_|$)`), "job_title", genJobTitle},
	{regexp.MustCompile(`(?i)(^|_)(description|bio|about|summary)(_|$)`), "description", genDescription},
	{regexp.MustCompile(`(?i)^(latitude|lat)$`), "latitude", genLatitude},
	{regexp.MustCompile(`(?i)^(longitude|lng|lon)$`), "longitude", genLongitude},
	{regexp.MustCompile(`(?i)(^|_)(password|passwd|pwd|secret|api[-_]?key|token)(_|$)`), "password", genPassword},
}
func MatchColumnName(colName string, maxLen int, faker *gofakeit.Faker) (any, bool) {
	for _, p := range columnPatterns {
		if p.Pattern.MatchString(colName) {
			return p.Generate(faker, maxLen), true
		}
	}
	return nil, false
}
func genEmail(f *gofakeit.Faker, maxLen int) any {
	email := f.Email()
	if maxLen > 0 && len(email) > maxLen {
		email = f.LetterN(uint(min(5, maxLen-10))) + "@" + f.DomainName()
		if len(email) > maxLen {
			email = email[:maxLen]
		}
	}
	return email
}
func genUsername(f *gofakeit.Faker, maxLen int) any {
	username := f.Username()
	if maxLen > 0 && len(username) > maxLen {
		username = username[:maxLen]
	}
	return username
}
func genFirstName(f *gofakeit.Faker, maxLen int) any {
	name := f.FirstName()
	if maxLen > 0 && len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}
func genLastName(f *gofakeit.Faker, maxLen int) any {
	name := f.LastName()
	if maxLen > 0 && len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}
func genFullName(f *gofakeit.Faker, maxLen int) any {
	name := f.Name()
	if maxLen > 0 && len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}
func genPhone(f *gofakeit.Faker, maxLen int) any {
	phone := f.Phone()
	if maxLen > 0 && len(phone) > maxLen {
		phone = f.Numerify(strings.Repeat("#", min(10, maxLen)))
	}
	return phone
}
func genStreet(f *gofakeit.Faker, maxLen int) any {
	street := f.Street()
	if maxLen > 0 && len(street) > maxLen {
		street = street[:maxLen]
	}
	return street
}
func genCity(f *gofakeit.Faker, maxLen int) any {
	city := f.City()
	if maxLen > 0 && len(city) > maxLen {
		city = city[:maxLen]
	}
	return city
}
func genState(f *gofakeit.Faker, maxLen int) any {
	if maxLen > 0 && maxLen <= 3 {
		return f.StateAbr()
	}
	state := f.State()
	if maxLen > 0 && len(state) > maxLen {
		return f.StateAbr()
	}
	return state
}
func genCountry(f *gofakeit.Faker, maxLen int) any {
	if maxLen > 0 && maxLen <= 3 {
		return f.CountryAbr()
	}
	country := f.Country()
	if maxLen > 0 && len(country) > maxLen {
		return f.CountryAbr()
	}
	return country
}
func genZip(f *gofakeit.Faker, maxLen int) any {
	zip := f.Zip()
	if maxLen > 0 && len(zip) > maxLen {
		zip = zip[:maxLen]
	}
	return zip
}
func genURL(f *gofakeit.Faker, maxLen int) any {
	url := f.URL()
	if maxLen > 0 && len(url) > maxLen {
		url = "https://" + f.DomainName()
		if len(url) > maxLen {
			url = url[:maxLen]
		}
	}
	return url
}
func genIP(f *gofakeit.Faker, _ int) any {
	return f.IPv4Address()
}
func genCompany(f *gofakeit.Faker, maxLen int) any {
	company := f.Company()
	if maxLen > 0 && len(company) > maxLen {
		company = company[:maxLen]
	}
	return company
}
func genJobTitle(f *gofakeit.Faker, maxLen int) any {
	title := f.JobTitle()
	if maxLen > 0 && len(title) > maxLen {
		title = title[:maxLen]
	}
	return title
}
func genDescription(f *gofakeit.Faker, maxLen int) any {
	wordCount := 10
	if maxLen > 0 && maxLen < 50 {
		wordCount = 3
	}
	desc := f.LoremIpsumSentence(wordCount)
	if maxLen > 0 && len(desc) > maxLen {
		desc = desc[:maxLen]
	}
	return desc
}
func genLatitude(f *gofakeit.Faker, _ int) any {
	return f.Latitude()
}
func genLongitude(f *gofakeit.Faker, _ int) any {
	return f.Longitude()
}
func genPassword(f *gofakeit.Faker, maxLen int) any {
	length := 16
	if maxLen > 0 && maxLen < length {
		length = maxLen
	}
	return f.Password(true, true, true, true, false, length)
}