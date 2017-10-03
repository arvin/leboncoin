insert into announces(
    href,
    title,
    price,
    published_at,
    urgent,
    department,
    city,
    district
) values($1, $2, $3, $4, $5, $6, $7, $8)
on conflict do nothing;