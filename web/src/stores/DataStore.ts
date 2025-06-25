import { makeAutoObservable, runInAction } from 'mobx';
import { AmoCRMField, AmoCRMPipeline, DialerScheduler, DialerCampaign, DialerBucket } from '../types';
import { RootStore } from './RootStore';
import api from '../services/api';

export class DataStore {
    rootStore: RootStore;
    amocrmFields: AmoCRMField[] = [];
    amocrmPipelines: AmoCRMPipeline[] = [];
    dialerSchedulers: DialerScheduler[] = [];
    dialerCampaigns: DialerCampaign[] = [];
    dialerBuckets: DialerBucket[] = [];
    isLoading = false;
    error: string | null = null;

    constructor(rootStore: RootStore) {
        this.rootStore = rootStore;
        makeAutoObservable(this);
    }

    async loadAllData() {
        this.isLoading = true;
        this.error = null;
        try {
            const [fields, pipelines, schedulers, campaigns, buckets] = await Promise.all([
                api.get('/api/v1/amocrm/fields'),
                api.get('/api/v1/amocrm/pipelines'),
                api.get('/api/v1/dialer/schedulers'),
                api.get('/api/v1/dialer/campaigns'),
                api.get('/api/v1/dialer/buckets'),
            ]);

            runInAction(() => {
                this.amocrmFields = fields.data;
                this.amocrmPipelines = pipelines.data;
                this.dialerSchedulers = schedulers.data;
                this.dialerCampaigns = campaigns.data;
                this.dialerBuckets = buckets.data;
                this.isLoading = false;
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to load data';
                this.isLoading = false;
            });
        }
    }

    clear() {
        this.amocrmFields = [];
        this.amocrmPipelines = [];
        this.dialerSchedulers = [];
        this.dialerCampaigns = [];
        this.dialerBuckets = [];
        this.isLoading = false;
        this.error = null;
    }
}