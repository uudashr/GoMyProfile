/**
  This is the SQL script that will be used to initialize the database schema.
  We will evaluate you based on how well you design your database.
  1. How you design the tables.
  2. How you choose the data types and keys.
  3. How you name the fields.
  In this assignment we will use PostgreSQL as the database.
  */

CREATE TABLE users (
  id BYTEA PRIMARY KEY,
  phone_number VARCHAR(13) UNIQUE NOT NULL,
  full_name VARCHAR(60) NOT NULL,
  password_hash BYTEA NOT NULL,
  password_salt BYTEA NOT NULL
);
