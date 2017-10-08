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
where department in (?)
order by published_at desc;