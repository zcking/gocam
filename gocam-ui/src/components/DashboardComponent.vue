<template>
    <div class="gocam-dashboard">
        <h1>GoCam</h1>

        <button class="btn" v-bind:class="{ 'btn-success': this.status === 'on', 'btn-danger': this.status === 'off' }"
                v-on:click="togglePower()" id="PowerToggleBtn" type="button">Power {{ this.nstatus }}</button>

        <div class="row">
            <div class="col-lg-12">
                <!-- TODO: Implement and place a frame with live cam view in it here -->
                <iframe id="frame" class="cam-frame" src="http://localhost:4040/cam" width="80%"></iframe>
            </div>
        </div>
    </div>
</template>

<script>
    import axios from 'axios';
    export default {
        name: "DashboardComponent",
        data: function() {
            return {
                powerOn: false
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
            }
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
            }
        },
        created: function() {
            this.getPowerStatus()
        }
    }
</script>

<style scoped>
.cam-frame {
    margin-top: 10px;
    margin-left: auto;
    margin-right: auto;
}

/*#wrap { width: 400px; height: 400px; padding: 0; overflow: hidden; }*/
#frame { width: 660px; height: 500px; border: 1px solid black; }
#frame { zoom: 1.0; -moz-transform: scale(1.0); -moz-transform-origin: 0 0; }
</style>