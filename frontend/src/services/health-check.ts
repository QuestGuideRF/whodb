import { HealthActions } from '@/store/health';
import { reduxStore } from '@/store';
import { GetHealthDocument } from '@graphql';
import { graphqlClient } from '@/config/graphql-client';

type HealthResponse = {
    server: 'healthy' | 'error' | 'unavailable';
    database: 'healthy' | 'error' | 'unavailable';
};

type HealthCheckConfig = {
    initialInterval: number;
    maxInterval: number;
    backoffMultiplier: number;
};

const DEFAULT_CONFIG: HealthCheckConfig = {
    initialInterval: 5000,
    maxInterval: 60000,
    backoffMultiplier: 1.5,
};

class HealthCheckService {
    private intervalId: ReturnType<typeof setTimeout> | null = null;
    private currentInterval: number;
    private config: HealthCheckConfig;
    private consecutiveFailures: number = 0;
    private isRunning: boolean = false;
    private wasDown: boolean = false;
    private wasDatabaseDown: boolean = false;

    constructor(config: Partial<HealthCheckConfig> = {}) {
        this.config = { ...DEFAULT_CONFIG, ...config };
        this.currentInterval = this.config.initialInterval;
    }

    private async checkHealth(): Promise<HealthResponse | null> {
        try {
            const result = await graphqlClient.query({
                query: GetHealthDocument,
                fetchPolicy: 'network-only',
                errorPolicy: 'all',
            });

            if (result.error) {
                return null;
            }

            if (!result.data || !result.data.Health) {
                return null;
            }

            const health = result.data.Health;

            return {
                server: (health.Server?.toLowerCase() || 'unavailable') as 'healthy' | 'error' | 'unavailable',
                database: (health.Database?.toLowerCase() || 'unavailable') as 'healthy' | 'error' | 'unavailable',
            };
        } catch (error) {

            return null;
        }
    }

    private async performHealthCheck(): Promise<void> {
        const result = await this.checkHealth();

        if (result === null) {

            reduxStore.dispatch(HealthActions.setHealthStatus({
                server: 'error',
                database: 'unavailable',
            }));
            this.consecutiveFailures++;
            this.wasDown = true;
            this.wasDatabaseDown = true;
            this.increaseInterval();
        } else {

            const isServerHealthy = result.server === 'healthy';
            const isDatabaseHealthy = result.database === 'healthy';
            const isFullyHealthy = isServerHealthy && isDatabaseHealthy;

            const isDatabaseError = result.database === 'error';
            if (isDatabaseError) {
                this.wasDatabaseDown = true;
            }

            reduxStore.dispatch(HealthActions.setHealthStatus({
                server: isServerHealthy ? 'healthy' : 'error',
                database: result.database,
            }));

            const serverRecovered = this.wasDown && isServerHealthy;
            const databaseRecovered = this.wasDatabaseDown && isDatabaseHealthy;

            if (serverRecovered || databaseRecovered) {
                window.location.reload();
                return;
            }

            if (isFullyHealthy) {
                this.resetInterval();
            } else {
                this.consecutiveFailures++;
                this.increaseInterval();
            }
        }
    }

    private increaseInterval(): void {
        this.currentInterval = Math.min(
            this.currentInterval * this.config.backoffMultiplier,
            this.config.maxInterval
        );
    }

    private resetInterval(): void {
        if (this.consecutiveFailures > 0) {
            this.consecutiveFailures = 0;
            this.currentInterval = this.config.initialInterval;
            this.reschedule();
        }
    }

    private reschedule(): void {
        if (this.intervalId !== null) {
            clearTimeout(this.intervalId);
        }

        if (this.isRunning) {
            this.intervalId = setTimeout(() => {
                this.performHealthCheck();
                this.reschedule();
            }, this.currentInterval);
        }
    }

    start(): void {
        if (this.isRunning) {
            return;
        }

        this.isRunning = true;
        this.consecutiveFailures = 0;
        this.wasDown = false;
        this.wasDatabaseDown = false;
        this.currentInterval = this.config.initialInterval;

        this.reschedule();
    }

    stop(): void {
        this.isRunning = false;
        if (this.intervalId !== null) {
            clearTimeout(this.intervalId);
            this.intervalId = null;
        }
    }

    async forceCheck(): Promise<void> {
        await this.performHealthCheck();
    }
}

export const healthCheckService = new HealthCheckService();