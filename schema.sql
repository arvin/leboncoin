create table announces(
    id serial primary key,
    href text unique not null,
    title text not null,
    price int,
    published_at timestamp with time zone not null,
    urgent boolean not null,
    department text,
    city text,
    district text
);