
CREATE TABLE contacts (
  id serial primary key,
  account_id integer,
  created_at timestamptz DEFAULT current_timestamp,
  updated_at timestamptz DEFAULT NULL,
  first_name varchar(255) DEFAULT NULL,
  middle_name varchar(255) DEFAULT NULL,
  last_name varchar(255) DEFAULT NULL,
  email_address varchar(255) DEFAULT NULL
);

CREATE FUNCTION set_updated_at()
  RETURNS trigger as $$
  BEGIN
    NEW.updated_at := current_timestamp;
    RETURN NEW;
  END $$ language plpgsql;

CREATE TRIGGER contacts_updated_at
  BEFORE UPDATE OR INSERT ON contacts
  FOR EACH ROW
  EXECUTE PROCEDURE set_updated_at();
