import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, TextField, Heading, Flex, Text, TextArea, Select } from '@radix-ui/themes';
import { createJob } from '../services/jobService';

const JobCreation: React.FC = () => {
    const [jobName, setJobName] = useState('');
    const [jobDescription, setJobDescription] = useState('');
    const [cronExpression, setCronExpression] = useState('');
    const [priority, setPriority] = useState(1);
    const [maxRetries, setMaxRetries] = useState(0);
    const [timeout, setTimeout] = useState(0);
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const newJob = await createJob({
            name: jobName,
            description: jobDescription,
            cron_expression: cronExpression,
            priority,
            max_retries: maxRetries,
            timeout
        });
        navigate(`/job/${newJob.jobId}`);
    };

    return (
        <Flex direction="column" gap="4" style={{ padding: '20px', maxWidth: '500px', margin: '0 auto' }}>
            <Heading size="4">Create New Job</Heading>
            <form onSubmit={handleSubmit}>
                <Flex direction="column" gap="2">
                    <label htmlFor="jobName">
                        <Text size="2">Job Name</Text>
                    </label>
                    <TextField.Input
                        id="jobName"
                        value={jobName}
                        onChange={(e) => setJobName(e.target.value)}
                        placeholder="Enter job name"
                        required
                    />
                </Flex>
                <Flex direction="column" gap="2" style={{ marginTop: '10px' }}>
                    <label htmlFor="jobDescription">
                        <Text size="2">Job Description</Text>
                    </label>
                    <TextArea
                        id="jobDescription"
                        value={jobDescription}
                        onChange={(e) => setJobDescription(e.target.value)}
                        placeholder="Enter job description"
                        required
                    />
                </Flex>
                <Flex direction="column" gap="2" style={{ marginTop: '10px' }}>
                    <label htmlFor="cronExpression">
                        <Text size="2">Cron Expression</Text>
                    </label>
                    <TextField.Input
                        id="cronExpression"
                        value={cronExpression}
                        onChange={(e) => setCronExpression(e.target.value)}
                        placeholder="Enter cron expression"
                        required
                    />
                </Flex>
                <Flex direction="column" gap="2" style={{ marginTop: '10px' }}>
                    <label htmlFor="priority">
                        <Text size="2">Priority</Text>
                    </label>
                    <Select.Root value={priority.toString()} onValueChange={(value) => setPriority(parseInt(value))}>
                        <Select.Trigger />
                        <Select.Content>
                            {[1, 2, 3, 4, 5].map((p) => (
                                <Select.Item key={p} value={p.toString()}>{p}</Select.Item>
                            ))}
                        </Select.Content>
                    </Select.Root>
                </Flex>
                <Flex direction="column" gap="2" style={{ marginTop: '10px' }}>
                    <label htmlFor="maxRetries">
                        <Text size="2">Max Retries</Text>
                    </label>
                    <TextField.Input
                        id="maxRetries"
                        type="number"
                        value={maxRetries}
                        onChange={(e) => setMaxRetries(parseInt(e.target.value))}
                        placeholder="Enter max retries"
                        required
                    />
                </Flex>
                <Flex direction="column" gap="2" style={{ marginTop: '10px' }}>
                    <label htmlFor="timeout">
                        <Text size="2">Timeout (seconds)</Text>
                    </label>
                    <TextField.Input
                        id="timeout"
                        type="number"
                        value={timeout}
                        onChange={(e) => setTimeout(parseInt(e.target.value))}
                        placeholder="Enter timeout in seconds"
                        required
                    />
                </Flex>
                <Button type="submit" style={{ marginTop: '20px' }}>Create Job</Button>
            </form>
        </Flex>
    );
};

export default JobCreation;
