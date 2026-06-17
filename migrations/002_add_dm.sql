-- DM機能テーブル追加
USE hackathon;

CREATE TABLE IF NOT EXISTS dm_rooms (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user1_id        BIGINT   NOT NULL COMMENT 'smaller user_id',
    user2_id        BIGINT   NOT NULL COMMENT 'larger user_id',
    set_id          BIGINT   NULL     COMMENT 'related starter set (context)',
    last_message    TEXT     NULL,
    last_message_at DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_users (user1_id, user2_id),
    INDEX idx_user1 (user1_id),
    INDEX idx_user2 (user2_id),
    FOREIGN KEY (user1_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS dm_messages (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id     BIGINT   NOT NULL,
    sender_id   BIGINT   NOT NULL,
    body        TEXT     NOT NULL,
    is_read     BOOLEAN  NOT NULL DEFAULT FALSE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_room_created (room_id, created_at),
    FOREIGN KEY (room_id)   REFERENCES dm_rooms (id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users   (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
