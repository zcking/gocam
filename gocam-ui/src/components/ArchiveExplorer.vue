<template>
    <div class="archive-explorer">
        <div class="card">
            <div class="card-header">
                <span clas="card-title">Archives</span>
            </div>

            <div class="card-body">
                <ul class="list-group list-group-flush">
                    <li v-for="archive in archives" v-bind:key="archive.Name" class="list-group-item">
                        {{ archive.Name }} <a href="#" class="text-danger" v-on:click="deleteArchive(archive.Name)">&#x274C;</a>
                    </li>
                </ul>
            </div>

            <div class="card-footer">
                <button class="btn btn-outline-dark" type="button" v-on:click="fetchArchives">Refresh</button>
            </div>
        </div>
    </div>
</template>

<script>
    import axios from 'axios';

    export default {
        name: "ArchiveExplorer",
        data: function() {
            return {
                archives: []
            }
        },
        methods: {
            fetchArchives: function() {
                axios
                    .get("http://localhost:4040/api/archives")
                    .then(resp => this.archives = resp.data)
            },
            deleteArchive: function(archiveName) {
                this.$emit('add-alert', {
                    level: 'danger',
                    text: 'The feature to delete archives has not yet been implemented; cannot delete ' + archiveName
                })
            }
        },
        created: function() {
            this.fetchArchives()
        }
    }
</script>

<style scoped>

</style>