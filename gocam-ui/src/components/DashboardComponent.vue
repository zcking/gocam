<template>
    <div class="gocam-dashboard">
        <h1>GoCam</h1>

        <div class="alerts">
            <div v-for="alert in alerts" v-bind:class="alertClass(alert)" role="alert">
                {{ alert.text }}
            </div>
        </div>

        <button class="btn" v-bind:class="{ 'btn-success': this.status === 'on', 'btn-danger': this.status === 'off' }"
                v-on:click="togglePower()" id="PowerToggleBtn" type="button">Power {{ this.nstatus }}</button>

        <div class="row content">
            <div class="col-lg-8">
                <iframe id="frame" class="cam-frame" src="http://localhost:4040/cam" width="80%"></iframe>
            </div>

            <div class="col-lg-4">
                <archive-explorer v-on:add-alert="addAlert($event)"/>
            </div>
        </div>
    </div>
</template>

<script>
    import ArchiveExplorer from './ArchiveExplorer'
    import axios from 'axios';
    export default {
        name: "DashboardComponent",
        components: {
            ArchiveExplorer,
        },
        data: function() {
            return {
                powerOn: false,
                alerts: []
            }
        },
        computed: {
            status: function() {
                if (this.powerOn) {
                    return "on"
                } else {
                    return "off"
                }
            },
            nstatus: function() {
                if (this.powerOn) {
                    return "off"
                } else {
                    return "on"
                }
            },
        },
        methods: {
            getPowerStatus: function() {
                axios('http://localhost:4040/api/power', {
                    method: 'GET',
                }).then(response => {
                    this.powerOn = response.data.PowerOn
                })
            },
            setPower: function(power) {
                var flag = "off";
                if (power) {
                    flag = "on";
                }
                axios
                    .get("http://localhost:4040/api/power/" + flag)
                    .then(response => (this.powerOn = response.data.PowerOn))
            },
            togglePower: function() {
                this.setPower(!this.powerOn);
            },
            alertClass: function(alert) {
                return "alert alert-" + alert.level
            },
            addAlert: function(alert) {
                alert.id = Object.keys(this.alerts).length;
                this.alerts.push(alert)
            }
        },
        created: function() {
            this.getPowerStatus()
        }
    }
</script>

<style scoped>
.cam-frame {
    margin-left: auto;
    margin-right: auto;
}

.content {
    margin-top: 10px;
}

/*#wrap { width: 400px; height: 400px; padding: 0; overflow: hidden; }*/
#frame { width: 660px; height: 500px; border: 1px solid black; }
#frame { zoom: 1.0; -moz-transform: scale(1.0); -moz-transform-origin: 0 0; }
</style>