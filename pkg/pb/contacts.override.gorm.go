package pb

// BeforeToORM ...
func (m *Contact) BeforeToORM(c *ContactORM) {
	if m.PrimaryEmail != "" {
		primary := &EmailORM{Address: m.PrimaryEmail, IsPrimary: true}
		for _, mail := range m.Emails {
			if mail.Address == primary.Address {
				return
			}
		}
		c.Emails = append(c.Emails, primary)
	}
}

// AfterToPB move the primary e-mail to the special field
func (m *ContactORM) AfterToPB(c *Contact) {
	if len(m.Emails) == 0 {
		return
	}
	// find the primary e-mail in list of e-mails from DB
	for _, addr := range m.Emails {
		if addr != nil && addr.IsPrimary {
			c.PrimaryEmail = addr.Address
			break
		}
	}
}
