import { format } from 'date-fns';
import React, { useState, useEffect, useCallback } from 'react';
import { Table, Heading, Flex, Text, Button } from '@radix-ui/themes';
import { Job, Execution } from '../types';
import { getExecutionHistory } from '../services/executionService';

const ExecutionHistoryList: React.FC<{ job: Job }> = ({ job }) => {
    const [executions, setExecutions] = useState<Execution[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchExecutionHistory = useCallback(async () => {
        if (!job.id) return;

        try {
            setLoading(true);
            const history = await getExecutionHistory(job.id);
            setExecutions(history);
            setError(null);
        } catch (err) {
            setError('Failed to fetch execution history');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [job.id]);

    useEffect(() => {
        fetchExecutionHistory();
    }, [fetchExecutionHistory]);

    const handleRefresh = () => {
        fetchExecutionHistory();
    };

    if (loading) {
        return <Text>Loading execution history...</Text>;
    }

    if (error) {
        return <Text color="red">{error}</Text>;
    }

    return (
        <>
            <Flex justify="between" align="center" style={{ marginBottom: '20px' }}>
                <Heading size="4">Execution History</Heading>
                <Button onClick={handleRefresh}>Refresh</Button>
            </Flex>
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.ColumnHeaderCell>Execution ID</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Status</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Start Time</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>End Time</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Result</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Error</Table.ColumnHeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {executions.map((execution) => (
                        <Table.Row key={execution.id}>
                            <Table.Cell>{execution.id}</Table.Cell>
                            <Table.Cell>{execution.status}</Table.Cell>
                            <Table.Cell>{format(new Date(execution.startTime), 'yyyy-MM-dd HH:mm:ss')}</Table.Cell>
                            <Table.Cell>{execution.endTime ? format(new Date(execution.endTime), 'yyyy-MM-dd HH:mm:ss') : 'In Progress'}</Table.Cell>
                            <Table.Cell>{execution.result || '-'}</Table.Cell>
                            <Table.Cell>{execution.error || '-'}</Table.Cell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table.Root>
            {executions.length === 0 && (
                <Flex justify="center" style={{ marginTop: '20px' }}>
                    <Text>No execution history found for this job.</Text>
                </Flex>
            )}
        </>
    );
};

export default React.memo(ExecutionHistoryList);
