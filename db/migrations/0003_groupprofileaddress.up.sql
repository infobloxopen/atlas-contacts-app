CREATE TABLE profiles
(
  id serial primary key,
  account_id text,
  name text,
  notes text
);

CREATE TABLE groups
(
  id serial primary key,
  account_id text,
  name text,
  notes text,
  profile_id int REFERENCES profiles(id) ON DELETE SET NULL
);

CREATE TABLE group_contacts
(
  group_id int REFERENCES groups(id) ON DELETE CASCADE,
  contact_id int REFERENCES contacts(id) ON DELETE CASCADE,
  primary key (group_id, contact_id)
);

CREATE TABLE addresses
(
  account_id text,
  address text,
  city text,
  country text,
  home_address_contact_id int REFERENCES contacts(id) ON DELETE CASCADE,
  state text,
  work_address_contact_id int REFERENCES contacts(id) ON DELETE CASCADE,
  zip text
);

ALTER TABLE contacts ADD COLUMN notes text;
ALTER TABLE contacts ADD COLUMN profile_id int REFERENCES profiles(id) ON DELETE SET NULL;
