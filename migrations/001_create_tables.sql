-- ホビーリレー データベーススキーマ
-- MySQL 8.0

CREATE DATABASE IF NOT EXISTS hobby_relay DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE hobby_relay;

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    display_name   VARCHAR(100)   NOT NULL,
    email          VARCHAR(255)   NOT NULL UNIQUE,
    password_hash  VARCHAR(255)   NOT NULL,
    avatar_url     VARCHAR(500)   NOT NULL DEFAULT '',
    rating_average DECIMAL(3, 2)  NOT NULL DEFAULT 0.00,
    rating_count   INT            NOT NULL DEFAULT 0,
    created_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 趣味カテゴリテーブル
CREATE TABLE IF NOT EXISTS hobby_categories (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500) NOT NULL DEFAULT '',
    icon_name   VARCHAR(100) NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_slug (slug),
    INDEX idx_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 趣味テーブル
CREATE TABLE IF NOT EXISTS hobbies (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    category_id BIGINT       NOT NULL,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category_id (category_id),
    INDEX idx_slug (slug),
    FOREIGN KEY (category_id) REFERENCES hobby_categories (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- スターターセットテーブル
CREATE TABLE IF NOT EXISTS starter_sets (
    id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    seller_id           BIGINT         NOT NULL,
    hobby_id            BIGINT         NOT NULL DEFAULT 0,
    category_id         BIGINT         NOT NULL DEFAULT 0,
    title               VARCHAR(255)   NOT NULL DEFAULT '',
    description         TEXT           NULL,
    price               INT            NOT NULL DEFAULT 0,
    status              VARCHAR(20)    NOT NULL DEFAULT 'draft'
                            COMMENT 'draft, on_sale, reserved, sold, hidden',
    beginner_score      TINYINT        NOT NULL DEFAULT 0 COMMENT '1-5',
    readiness_score     TINYINT        NOT NULL DEFAULT 0 COMMENT '0-100',
    value_score         TINYINT        NOT NULL DEFAULT 0,
    estimated_new_price INT            NOT NULL DEFAULT 0,
    previous_owner_note TEXT           NULL,
    startable_summary   TEXT           NULL,
    published_at        DATETIME       NULL,
    created_at          DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_seller_id (seller_id),
    INDEX idx_hobby_id (hobby_id),
    INDEX idx_category_id (category_id),
    INDEX idx_status (status),
    INDEX idx_status_readiness (status, readiness_score),
    INDEX idx_status_published (status, published_at),
    INDEX idx_price (price),
    FOREIGN KEY (seller_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- セット画像テーブル
CREATE TABLE IF NOT EXISTS set_images (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    starter_set_id BIGINT       NOT NULL,
    image_url      VARCHAR(500) NOT NULL,
    sort_order     INT          NOT NULL DEFAULT 0,
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_starter_set_id (starter_set_id),
    INDEX idx_sort_order (starter_set_id, sort_order),
    FOREIGN KEY (starter_set_id) REFERENCES starter_sets (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- セットアイテムテーブル
CREATE TABLE IF NOT EXISTS set_items (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    starter_set_id  BIGINT       NOT NULL,
    name            VARCHAR(255) NOT NULL,
    condition_label VARCHAR(20)  NOT NULL DEFAULT 'unknown'
                        COMMENT 'new, like_new, good, fair, unknown',
    quantity        INT          NOT NULL DEFAULT 1,
    is_essential    BOOLEAN      NOT NULL DEFAULT FALSE,
    note            TEXT         NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_starter_set_id (starter_set_id),
    FOREIGN KEY (starter_set_id) REFERENCES starter_sets (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- おすすめアイテムテーブル
CREATE TABLE IF NOT EXISTS recommended_items (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    starter_set_id  BIGINT       NOT NULL,
    name            VARCHAR(255) NOT NULL,
    importance      VARCHAR(20)  NOT NULL DEFAULT 'recommended'
                        COMMENT 'required, recommended, nice_to_have',
    reason          TEXT         NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_starter_set_id (starter_set_id),
    FOREIGN KEY (starter_set_id) REFERENCES starter_sets (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- お気に入りテーブル
CREATE TABLE IF NOT EXISTS favorites (
    user_id         BIGINT   NOT NULL,
    starter_set_id  BIGINT   NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, starter_set_id),
    INDEX idx_user_id (user_id),
    INDEX idx_starter_set_id (starter_set_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (starter_set_id) REFERENCES starter_sets (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- トランザクションテーブル
CREATE TABLE IF NOT EXISTS transactions (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    starter_set_id  BIGINT      NOT NULL,
    buyer_id        BIGINT      NOT NULL,
    seller_id       BIGINT      NOT NULL,
    price           INT         NOT NULL DEFAULT 0,
    status          VARCHAR(30) NOT NULL DEFAULT 'reserved'
                        COMMENT 'reserved, handover_waiting, shipped, received, completed, cancelled',
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_starter_set_id (starter_set_id),
    INDEX idx_buyer_id (buyer_id),
    INDEX idx_seller_id (seller_id),
    INDEX idx_status (status),
    FOREIGN KEY (starter_set_id) REFERENCES starter_sets (id),
    FOREIGN KEY (buyer_id) REFERENCES users (id),
    FOREIGN KEY (seller_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- AIチャットセッションテーブル
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT      NOT NULL,
    starter_set_id  BIGINT      NULL,
    session_type    VARCHAR(30) NOT NULL DEFAULT 'listing_support'
                        COMMENT 'listing_support, set_question, search, start_plan',
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                        COMMENT 'active, completed, cancelled',
    progress_step   INT         NOT NULL DEFAULT 1,
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_starter_set_id (starter_set_id),
    INDEX idx_session_type (session_type),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- AIメッセージテーブル
CREATE TABLE IF NOT EXISTS ai_messages (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id  BIGINT      NOT NULL,
    sender      VARCHAR(20) NOT NULL DEFAULT 'user'
                    COMMENT 'user, assistant, system',
    message     TEXT        NOT NULL,
    metadata    TEXT        NULL,
    created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_session_created (session_id, created_at),
    FOREIGN KEY (session_id) REFERENCES ai_chat_sessions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- スタートプランテーブル
CREATE TABLE IF NOT EXISTS start_plans (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    transaction_id  BIGINT       NOT NULL,
    starter_set_id  BIGINT       NOT NULL,
    user_id         BIGINT       NOT NULL,
    title           VARCHAR(255) NOT NULL DEFAULT '',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_user_id (user_id),
    UNIQUE KEY uq_transaction_id (transaction_id),
    FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- スタートプランステップテーブル
CREATE TABLE IF NOT EXISTS start_plan_steps (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    start_plan_id BIGINT       NOT NULL,
    day_no        INT          NOT NULL DEFAULT 1,
    title         VARCHAR(255) NOT NULL DEFAULT '',
    body          TEXT         NULL,
    INDEX idx_start_plan_id (start_plan_id),
    UNIQUE KEY uq_plan_day (start_plan_id, day_no),
    FOREIGN KEY (start_plan_id) REFERENCES start_plans (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
