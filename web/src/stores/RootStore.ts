import { makeAutoObservable } from 'mobx';
import { AuthStore } from './AuthStore';
import { FlowStore } from './FlowStore';
import { DataStore } from './DataStore';

export class RootStore {
    authStore: AuthStore;
    flowStore: FlowStore;
    dataStore: DataStore;

    constructor() {
        this.authStore = new AuthStore(this);
        this.flowStore = new FlowStore(this);
        this.dataStore = new DataStore(this);
        makeAutoObservable(this);
    }
}