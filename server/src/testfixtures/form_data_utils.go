package testfixtures

import "github.com/dxe/adb/model"

type InternationalFormDataBuilder struct {
	formData model.InternationalFormData
}

func NewInternationalFormDataBuilder() *InternationalFormDataBuilder {
	return &InternationalFormDataBuilder{
		formData: model.InternationalFormData{
			ID:          1,
			FirstName:   "DefaultFirstName",
			LastName:    "DefaultLastName",
			Email:       "default@example.com",
			Phone:       "1234567890",
			Interest:    "DefaultInterest",
			Involvement: "DefaultInvolvement",
			City:        "DefaultCity",
			State:       "DefaultState",
			Country:     "DefaultCountry",
			Lat:         37.7749, Lng: -122.4194},
	}
}

func (b *InternationalFormDataBuilder) WithID(id int) *InternationalFormDataBuilder {
	b.formData.ID = id
	return b
}

func (b *InternationalFormDataBuilder) WithFirstName(firstName string) *InternationalFormDataBuilder {
	b.formData.FirstName = firstName
	return b
}

func (b *InternationalFormDataBuilder) WithLastName(lastName string) *InternationalFormDataBuilder {
	b.formData.LastName = lastName
	return b
}

func (b *InternationalFormDataBuilder) WithEmail(email string) *InternationalFormDataBuilder {
	b.formData.Email = email
	return b
}

func (b *InternationalFormDataBuilder) WithPhone(phone string) *InternationalFormDataBuilder {
	b.formData.Phone = phone
	return b
}

func (b *InternationalFormDataBuilder) WithInterest(interest string) *InternationalFormDataBuilder {
	b.formData.Interest = interest
	return b
}

func (b *InternationalFormDataBuilder) WithInvolvement(involvement string) *InternationalFormDataBuilder {
	b.formData.Involvement = involvement
	return b
}

func (b *InternationalFormDataBuilder) WithCity(city string) *InternationalFormDataBuilder {
	b.formData.City = city
	return b
}

func (b *InternationalFormDataBuilder) WithState(state string) *InternationalFormDataBuilder {
	b.formData.State = state
	return b
}

func (b *InternationalFormDataBuilder) WithCountry(country string) *InternationalFormDataBuilder {
	b.formData.Country = country
	return b
}

func (b *InternationalFormDataBuilder) WithLat(lat float64) *InternationalFormDataBuilder {
	b.formData.Lat = lat
	return b
}

func (b *InternationalFormDataBuilder) WithLng(lng float64) *InternationalFormDataBuilder {
	b.formData.Lng = lng
	return b
}

func (b *InternationalFormDataBuilder) Build() model.InternationalFormData {
	return b.formData
}
