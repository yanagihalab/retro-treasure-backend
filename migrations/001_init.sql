CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE player_status (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    level INT NOT NULL DEFAULT 1,
    exp INT NOT NULL DEFAULT 0,
    stamina INT NOT NULL DEFAULT 20,
    max_stamina INT NOT NULL DEFAULT 20,
    coins INT NOT NULL DEFAULT 0,
    gems INT NOT NULL DEFAULT 0,
    total_explorations INT NOT NULL DEFAULT 0,
    last_stamina_recovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE areas (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    required_level INT NOT NULL DEFAULT 1,
    stamina_cost INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    rarity INT NOT NULL DEFAULT 1,
    item_type VARCHAR(30) NOT NULL,
    sell_price INT NOT NULL DEFAULT 0,
    is_encyclopedia_target BOOLEAN NOT NULL DEFAULT TRUE,
    is_event_limited BOOLEAN NOT NULL DEFAULT FALSE,
    icon_path TEXT
);

CREATE TABLE area_drop_tables (
    id BIGSERIAL PRIMARY KEY,
    area_id BIGINT NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    item_id BIGINT REFERENCES items(id) ON DELETE CASCADE,
    drop_weight INT NOT NULL DEFAULT 1,
    exp_reward INT NOT NULL DEFAULT 0,
    coin_reward INT NOT NULL DEFAULT 0,
    event_type VARCHAR(30),
    message TEXT,
    UNIQUE(area_id, item_id, event_type)
);

CREATE TABLE user_items (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE encyclopedia_entries (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    first_obtained_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE exploration_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    area_id BIGINT NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    consumed_stamina INT NOT NULL,
    gained_exp INT NOT NULL DEFAULT 0,
    gained_coins INT NOT NULL DEFAULT 0,
    result_type VARCHAR(30) NOT NULL,
    result_item_id BIGINT REFERENCES items(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE login_bonus_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bonus_date DATE NOT NULL,
    reward_type VARCHAR(30) NOT NULL,
    reward_value INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, bonus_date)
);

CREATE TABLE notices (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
