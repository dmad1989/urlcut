-- +goose Up
-- +goose StatementBegin
CREATE TABLE if not exists  public.USERS
(
    "USERID" bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 3 MINVALUE 3 MAXVALUE 1000000 ),
    CONSTRAINT "PK_USERS" PRIMARY KEY ("USERID")
)
TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.USERS
    OWNER to postgres;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.USERS;
-- +goose StatementEnd
