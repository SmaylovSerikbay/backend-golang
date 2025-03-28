-- Создаем таблицу для сообщений чата
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    ride_id INTEGER NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_chat_messages_ride FOREIGN KEY (ride_id) REFERENCES rides(id),
    CONSTRAINT fk_chat_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Создаем таблицу для отслеживания прочитанных сообщений
CREATE TABLE IF NOT EXISTS chat_message_reads (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    ride_id INTEGER NOT NULL REFERENCES rides(id),
    read_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_chat_message_reads_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_chat_message_reads_ride FOREIGN KEY (ride_id) REFERENCES rides(id)
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_chat_messages_ride_id ON chat_messages(ride_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

CREATE INDEX idx_chat_message_reads_user_id ON chat_message_reads(user_id);
CREATE INDEX idx_chat_message_reads_ride_id ON chat_message_reads(ride_id);
CREATE INDEX idx_chat_message_reads_read_at ON chat_message_reads(read_at); 