# Shorty is an URL shortner with unique ids.

## Theoretical limit:
`math.MaxUint32` ids per second per machine up to year 2059.

---

## Example Usage:
```
// on machine 1
store, err := shorty.NewBoltStore(db)
check(err)
s, err := short.New(store, 1)
check(err)
id, err := s.GenerateID("http://google.com")
check(err)
....

// on machine 2
store, err := shorty.NewBoltStore(db)
check(err)
s, err := short.New(store, 2)
check(err)
id, err := s.GenerateID("http://google.com")
check(err)

// on machine N

store, err := shorty.NewBoltStore(db)
check(err)
s, err := short.New(store, N)
check(err)
id, err := s.GenerateID("http://google.com")
check(err)

// on router
id, err := shorty.IDFromString(ctx.Param("id"))
check(err)
redirect internally based on id.MachineID()

```
