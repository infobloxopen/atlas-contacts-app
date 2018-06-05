
CREATE TABLE emails (
  id serial primary key,
  created_at timestamptz DEFAULT current_timestamp,
  updated_at timestamptz DEFAULT NULL,
  is_primary boolean DEFAULT false,
  address varchar(255) UNIQUE DEFAULT NULL,
  account_id varchar(255),
  contact_id int REFERENCES contacts(id) ON DELETE CASCADE
);

CREATE TRIGGER emails_updated_at
  BEFORE UPDATE OR INSERT ON emails
  FOR EACH ROW
  EXECUTE PROCEDURE set_updated_at();

INSERT INTO emails (address, is_primary, contact_id)
  SELECT email_address, true, id FROM contacts;

ALTER TABLE contacts DROP COLUMN email_address;
