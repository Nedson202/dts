import React, { useEffect, useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import * as Form from '@radix-ui/react-form';
import { Button } from './Button';
import { Job } from '../types';

interface JobEditDialogProps {
    job: Job | null;
    onSave: (updatedJob: Job) => void;
    onClose: () => void;
}

export function JobEditDialog({ job, onSave, onClose }: JobEditDialogProps) {
    const [editedJob, setEditedJob] = useState<Job | null>(job);

    useEffect(() => {
        setEditedJob(job);
    }, [job]);

    if (!editedJob) return null;

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSave(editedJob);
        onClose();
    };

    return (
        <Dialog.Root open={!job} onOpenChange={onClose}>
            <Dialog.Portal>
                <Dialog.Overlay className="fixed inset-0 bg-black/50" />
                <Dialog.Content className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-white p-6 rounded-lg shadow-lg">
                    <Dialog.Title className="text-lg font-bold mb-4">Edit Job</Dialog.Title>
                    <Form.Root onSubmit={handleSubmit}>
                        <Form.Field name="name">
                            <Form.Label>Name</Form.Label>
                            <Form.Control asChild>
                                <input
                                    type="text"
                                    value={editedJob.name}
                                    onChange={(e) => setEditedJob({ ...editedJob, name: e.target.value })}
                                    className="w-full p-2 border rounded"
                                />
                            </Form.Control>
                        </Form.Field>
                        <Form.Field name="description" className="mt-4">
                            <Form.Label>Description</Form.Label>
                            <Form.Control asChild>
                                <textarea
                                    value={editedJob.description}
                                    onChange={(e) => setEditedJob({ ...editedJob, description: e.target.value })}
                                    className="w-full p-2 border rounded"
                                />
                            </Form.Control>
                        </Form.Field>
                        <div className="mt-6 flex justify-end space-x-2">
                            <Button onClick={onClose} variant="secondary">Cancel</Button>
                            <Button type="submit">Save Changes</Button>
                        </div>
                    </Form.Root>
                </Dialog.Content>
            </Dialog.Portal>
        </Dialog.Root>
    );
}
