// GENERATED CODE! DO NOT EDIT.

package model

func (x *AlreadyExistsError) Reset() {
	if x == nil {
		return
	}

	x.ShortURL = ""
	x.Err = nil
}

func (x *Link) Reset() {
	if x == nil {
		return
	}

	if r, ok := any(x.UUID).(interface{ Reset() }); ok {
		r.Reset()
	}
	x.ShortURL = ""
	x.OriginalURL = ""
	x.CorrelationID = ""
	x.UserID = ""
	x.IsDeleted = false
}
