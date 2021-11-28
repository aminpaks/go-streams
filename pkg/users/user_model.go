package users

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/aminpaks/go-streams/pkg/merrors"
)

var emailRegex = regexp.MustCompile(`^([a-z0-9][-a-z0-9_\+\.]*[a-z0-9])@([a-z0-9][-a-z0-9\.]*[a-z0-9]\.(arpa|root|aero|biz|cat|com|coop|edu|gov|info|int|jobs|mil|mobi|museum|name|net|org|pro|tel|travel|ac|ad|ae|af|ag|ai|al|am|an|ao|aq|ar|as|at|au|aw|ax|az|ba|bb|bd|be|bf|bg|bh|bi|bj|bm|bn|bo|br|bs|bt|bv|bw|by|bz|ca|cc|cd|cf|cg|ch|ci|ck|cl|cm|cn|co|cr|cu|cv|cx|cy|cz|de|dj|dk|dm|do|dz|ec|ee|eg|er|es|et|eu|fi|fj|fk|fm|fo|fr|ga|gb|gd|ge|gf|gg|gh|gi|gl|gm|gn|gp|gq|gr|gs|gt|gu|gw|gy|hk|hm|hn|hr|ht|hu|id|ie|il|im|in|io|iq|ir|is|it|je|jm|jo|jp|ke|kg|kh|ki|km|kn|kr|kw|ky|kz|la|lb|lc|li|lk|lr|ls|lt|lu|lv|ly|ma|mc|md|mg|mh|mk|ml|mm|mn|mo|mp|mq|mr|ms|mt|mu|mv|mw|mx|my|mz|na|nc|ne|nf|ng|ni|nl|no|np|nr|nu|nz|om|pa|pe|pf|pg|ph|pk|pl|pm|pn|pr|ps|pt|pw|py|qa|re|ro|ru|rw|sa|sb|sc|sd|se|sg|sh|si|sj|sk|sl|sm|sn|so|sr|st|su|sv|sy|sz|tc|td|tf|tg|th|tj|tk|tl|tm|tn|to|tp|tr|tt|tv|tw|tz|ua|ug|uk|um|us|uy|uz|va|vc|ve|vg|vi|vn|vu|wf|ws|ye|yt|yu|za|zm|zw)|([0-9]{1,3}\.{3}[0-9]{1,3}))$`)

func NewUser(name, email string) *User {
	return &User{
		Id:    uuid.New(),
		Name:  name,
		Email: email,
	}
}

type User struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func (u *User) WithId(id uuid.UUID) *User {
	u.Id = id
	return u
}

func (u *User) Bytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *User) String() string {
	return string(u.Bytes())
}

func (u *User) Validate() error {
	errors := merrors.NewMerrors()
	if u.Name == "" {
		errors.AddCustom(merrors.BuildFiledValidationError("name", "Must be provided"))
	} else if len(u.Name) < 3 {
		errors.AddCustom(merrors.BuildFiledValidationError("name", "Must be at least 3 characters"))
	}
	if u.Email == "" {
		errors.AddCustom(merrors.BuildFiledValidationError("email", "Must be provided"))
	} else if !emailRegex.MatchString(u.Email) {
		errors.AddCustom(merrors.BuildFiledValidationError("email", fmt.Sprintf("Value '%s' is not a valid email address", u.Email)))
	}
	if errors.Has() {
		return errors
	}
	return nil
}

func ParseUser(v string) (*User, error) {
	var user User
	err := json.Unmarshal([]byte(v), &user)
	return &user, err
}

func ParseUserWithOptionalId(v []byte) (*User, error) {
	type userCreation struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var user *userCreation = &userCreation{}
	err := json.Unmarshal(v, user)
	if err != nil {
		return nil, err
	}

	return NewUser(user.Name, user.Email), nil
}
