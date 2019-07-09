-- DB -- 
create database s3db_01;
-- TABLE DEFINITIONS --
create table files_stored (
    id serial primary key,
    name text not null unique,
    remote_addr text not null,
    application text not null,
    bucket text not null,
    created timestamp not null default now()
);