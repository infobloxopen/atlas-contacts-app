package pb

// BeforeToORM will add the primary e-mail to the list of e-mails if it isn't
// present already
func (m *Contact) BeforeToORM(c *ContactORM) {
	if m.PrimaryEmail != "" {
		for _, mail := range m.Emails {
			if mail.Address == m.PrimaryEmail {
				return
			}
		}
		c.Emails = append(c.Emails, &EmailORM{Address: m.PrimaryEmail, IsPrimary: true})
	}
}

// AfterToPB copies the primary e-mail address from the DB to the special PB field
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
