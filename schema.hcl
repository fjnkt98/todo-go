table "items" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = text
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}
schema "public" {
  comment = "standard public schema"
}
