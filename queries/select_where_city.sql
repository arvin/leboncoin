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
where city in (?)
order by published_at desc;