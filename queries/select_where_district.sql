select
    href,
    title,
    price,
    published_at,
    urgent,
    department,
    city,
    district
from announces
where district in (?)
order by published_at desc;