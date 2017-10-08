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
where department in (?) or city in (?) or district in (?)
order by published_at desc;