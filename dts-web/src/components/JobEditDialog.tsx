import React, { useState, useEffect } from 'react';
import { Dialog, Button, Flex, TextField, Text } from '@radix-ui/themes';
import { JobEdit } from '../types';

interface JobEditDialogProps {
    isOpen: boolean;
    onClose: () => void;
    job: JobEdit | null;
    onSave: (updatedJob: JobEdit) => void;
}

export const JobEditDialog: React.FC<JobEditDialogProps> = ({ isOpen, onClose, job, onSave }) => {
    const [editedJob, setEditedJob] = useState<JobEdit | null>(null);

    useEffect(() => {
        console.log(job);
        if (job) {
            setEditedJob({
                id: job.id,
                name: job.name,
                description: job.description,
                cronExpression: job.cronExpression,
                metadata: job.metadata,
                priority: job.priority,
                maxRetries: job.maxRetries,
                timeout: job.timeout,
            });
        }
    }, [job]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (editedJob) {
            setEditedJob({ ...editedJob, [e.target.name]: e.target.value });
        }
    };

    const handleSave = () => {
        if (editedJob) {
            onSave(editedJob);
        }
        onClose();
    };

    if (!editedJob) return null;

    return (
        <Dialog.Root open={isOpen} onOpenChange={onClose}>
            <Dialog.Content style={{ maxWidth: 450 }}>
                <Dialog.Title>Edit Job</Dialog.Title>
                <Flex direction="column" gap="3">
                    <label>
                        <Text as="div" size="2" mb="1" weight="bold">
                            Name
                        </Text>
                        <TextField.Input
                            name="name"
                            value={editedJob.name}
                            onChange={handleInputChange}
                        />
                    </label>
                    <label>
                        <Text as="div" size="2" mb="1" weight="bold">
                            Description
                        </Text>
                        <TextField.Input
                            name="description"
                            value={editedJob.description}
                            onChange={handleInputChange}
                        />
                    </label>
                    <label>
                        <Text as="div" size="2" mb="1" weight="bold">
                            Cron Expression
                        </Text>
                        <TextField.Input
                            name="cronExpression"
                            value={editedJob.cronExpression}
                            onChange={handleInputChange}
                        />
                    </label>
                </Flex>
                <Flex gap="3" mt="4" justify="end">
                    <Dialog.Close>
                        <Button variant="soft" color="gray">
                            Cancel
                        </Button>
                    </Dialog.Close>
                    <Button onClick={handleSave}>Save Changes</Button>
                </Flex>
            </Dialog.Content>
        </Dialog.Root>
    );
};
