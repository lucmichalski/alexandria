/******************************
**	File:   mainqueries.sql
**	Name:	Database migrations scripts
**	Desc:	Main database migrations scripts for Author microservice
**	Auth:	Alonso R
**	Lic:	MIT	
**	Date:	2020-04-14
*******************************/

CREATE DATABASE 'alexandria/author';
SET search_path TO 'alexandria/author';

CREATE TYPE STATUS_ENUM AS ENUM(
    'STATUS_PENDING',
    'STATUS_PUBLISHED'
);

CREATE TABLE IF NOT EXISTS AUTHOR(
	ID 				BIGINT NOT NULL PRIMARY KEY,
	EXTERNAL_ID 	VARCHAR(128) NOT NULL UNIQUE,
	FIRST_NAME 		VARCHAR(255) NOT NULL,
	LAST_NAME 		VARCHAR(255) NOT NULL,
	DISPLAY_NAME 	VARCHAR(255) NOT NULL,
	BIRTH_DATE 		DATE NOT NULL DEFAULT CURRENT_DATE,
	OWNER_ID        VARCHAR(128) DEFAULT NULL,
	CREATE_TIME 	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UPDATE_TIME 	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	DELETE_TIME 	TIMESTAMP DEFAULT NULL,
	STATUS          STATUS_ENUM DEFAULT 'STATUS_PENDING',
	METADATA		TEXT DEFAULT NULL,
	DELETED 		BOOL DEFAULT FALSE
);