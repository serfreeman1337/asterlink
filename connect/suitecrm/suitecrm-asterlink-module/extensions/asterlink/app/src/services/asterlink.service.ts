import { inject, Injectable, OnDestroy, signal } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Subscription, BehaviorSubject } from 'rxjs';
import { AuthService } from 'core';

@Injectable({
    providedIn: 'root'
})
export class AsterlinkService implements OnDestroy {
    private authSub: Subscription;

    calls = signal([], {
        equal: () => false
    });

    ready = false;
    ready$ = new BehaviorSubject<boolean>(this.ready);
    
    token?: string;
    relations?: any;

    private endpointUrl?: string;
    
    private isUserLoggedIn = false;
    private worker?: SharedWorker;

    private http = inject(HttpClient);
    private authService = inject(AuthService);

    constructor() {
        this.authSub = this.authService.isUserLoggedIn.subscribe(val => {
            if (this.isUserLoggedIn == val)
                return;

            this.isUserLoggedIn = val;

            if (this.isUserLoggedIn) {
                this.onLogin();
            } else {
                this.reset();
            }
        });
    }


    ngOnDestroy(): void {
        this.authSub.unsubscribe();
        this.ready$.complete();
    }

    private onLogin() {
        const url = '?entryPoint=AsterLink';
        this.http.get(url, {
            headers: new HttpHeaders({
                'Accept': 'application/json'
            }),
            withCredentials: true
        }).subscribe((r: any) => {
            if (r) {
                if (r.token) {
                    this.token = r.token;
                    this.endpointUrl = r.endpoint_url + '/';
                } else {
                    this.endpointUrl = location.protocol + '//' + 
                            location.hostname + 
                            (location.port ? ':' + location.port : '' ) +
                            location.pathname + url + '&action=';
                }

                this.connectEvents();
                
                this.relations = r.relations;

                this.ready = true;
                this.ready$.next(true);
            } else {
                this.reset();
            }
        })
    }

    private reset() {
        this.token = undefined;
        this.endpointUrl = undefined;
        this.relations = undefined;

        if (this.worker) {
            this.worker.port.postMessage(false);
            this.worker = undefined;
        }

        this.ready = false;
        this.ready$.next(false);
    }

    private connectEvents() {
        this.worker = new SharedWorker('modules/AsterLink/javascript/asterlink.worker.js');

        this.worker.port.onmessage = e => {
            if (!e.data) {
                if (!e.data) { // Worker disconnected from stream.
                    this.calls.update(calls => {
                        calls.length = 0;
                        return calls;
                    });
                    return;
                }
            }

            const { show, data } = e.data;
            this.calls.update(calls => {
                const id = calls.findIndex(c => c.id == data.id);
                
                if (show) {
                    if (id == -1) {
                        calls.push(data);
                    } else {
                        Object.assign(calls[id], data);
                    }
                } else {
                    calls.splice(id, 1);
                }

                return calls;
            });
        };

        let url = this.endpointUrl + 'stream';

        if (this.token) {
            url += '?token=' + this.token;
        }

        this.worker.port.postMessage(url);
    }

    originate(phone: string) {
        const data = new URLSearchParams({ phone });

        let headers = new HttpHeaders({
            'Content-Type': 'application/x-www-form-urlencoded'
        });

        if (this.token) {
            headers = headers.append('X-Asterlink-Token', this.token);
        }

        return this.http.post(this.endpointUrl + 'originate', data, { headers });
    }
}