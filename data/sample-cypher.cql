MATCH
  (o:Office)<-[ran:`ran for`]-(c:Committee)<-[gave:`contributed to`]-(p:Person)
WITH
  c as c2, c as c1, p, ran, gave
WHERE
  c2.name <> c1.name AND ran.period = gave.period
RETURN p, c1