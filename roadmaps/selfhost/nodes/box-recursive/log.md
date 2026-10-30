# Log: box-recursive

- Nothing existed yet (no `Box`, "type contains itself" error for enums). Did the full work.
- Design: enums with fields are `{ int32_t tag; Q* u; }`, with Q (payload struct, kind letter `Q`, `payloadCName`) holding rc and a union of variant structs. Fieldless variants have `u == NULL`. The value struct is complete before Q, so `orderTypes` treats a boxed enum as having no inline dependencies.
- `Box<T>` is a native `BoxType`, C pointer to `{ rc, f_value }`, kind letter `B` (`boxCName`). `Box.new` is a namespace builtin; `.value` is handled in the checker's FieldExpr and the emitter.
- A user type named `Box` shadows the builtin (existing tests declare `struct Box<T>`).
- Sample with eval of depth 20000 overflowed the ASan stack in `eval` itself; lowered to 2000. List drop of 20000 is fine. A 1,000,000-deep drop crashes (recursive release).
