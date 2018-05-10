
ALTER TABLE contacts ADD COLUMN email_address varchar(255) DEFAULT NULL;

UPDATE contacts SET
  ( email_address ) = (
    SELECT address FROM emails WHERE emails.contact_id = contacts.id and emails.is_primary = true
  );

DROP TRIGGER emails_updated_at on emails;

DROP TABLE emails;
