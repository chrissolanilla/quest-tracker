-- difficulty weights (tiny lookup)
create table difficulty_weights (
  difficulty text primary key,
  weight real not null
);

insert into difficulty_weights(difficulty, weight) values
('easy', 1.0), ('medium', 2.0), ('hard', 3.0);

-- users
create table users (
  id text primary key,              -- asana user gid
  name text not null,
  avatar_url text,
  created_at timestamptz default now()
);

-- quests (mirror important task fields from asana)
create table quests (
  id text primary key,              -- asana task gid
  name text not null,
  difficulty text not null references difficulty_weights(difficulty),
  completed boolean not null default false,
  completed_by text references users(id),
  completed_at timestamptz
);

-- scores (materialized for fast leaderboard)
create table scores (
  user_id text primary key references users(id),
  points real not null default 0
);

-- oauth accounts: one per asana user
create table if not exists oauth_accounts (
  user_id text not null references users(id) on delete cascade,
  provider text not null,
  access_token text not null,
  refresh_token text not null,
  scope text,
  expires_at timestamptz not null,
  primary key (user_id, provider)
);

-- very simple sessions for dev (swap to redis later if you want)
create table if not exists sessions (
  id text primary key,
  user_id text not null references users(id) on delete cascade,
  created_at timestamptz default now()
);

