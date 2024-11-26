CREATE TABLE todo_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- UUID for Todo ID
    description TEXT NOT NULL,                    -- Task description
    due_date TIMESTAMP NOT NULL,                  -- Due date of the task
    file_id TEXT DEFAULT NULL,                    -- Reference to the S3 file
    created_at TIMESTAMP DEFAULT now(),           -- Creation timestamp
    updated_at TIMESTAMP DEFAULT now()            -- Update timestamp
);
