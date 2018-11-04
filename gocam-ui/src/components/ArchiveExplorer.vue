<template>
    <div class="archive-explorer">
        <div class="card">
            <div class="card-header">
                <span clas="card-title">Archives</span>
            </div>

            <div class="card-body scroll">
                <ul class="list-group list-group-flush">
                    <li v-for="archive in archives" v-bind:key="archive.Name" class="list-group-item">
                        <a v-bind:href="archiveHref(archive.Name)" title="Download" target="_blank">{{ archive.Name }}</a>&nbsp;
                        <a href="#" class="text-danger float-right" title="Delete" v-on:click="deleteArchive(archive.Name)">&#x274C;</a>
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
                axios({
                    method: 'get',
                    url: 'http://localhost:4040/api/archives/delete?archive=' + archiveName
                }).then(resp => {
                        if (resp.status === 200) {
                            this.fetchArchives();
                            this.$emit('add-alert', {
                                level: 'success',
                                text: archiveName + ' deleted.'
                            })
                        } else {
                            // eslint-disable-next-line
                            console.error(resp);
                            this.$emit('add-alert', {
                                level: 'danger',
                                text: resp.data
                            })
                        }
                    }).catch(err => {
                        console.error(err);
                        this.$emit('add-alert', {
                            level: 'danger',
                            text: err
                        })
                    });
            },
            archiveHref: function(archiveName) {
                return "http://localhost:4040/archives/" + archiveName
            }
        },
        created: function() {
            this.fetchArchives()
        }
    }
</script>

<style scoped>
    .scroll {
        max-height: 350px;
        overflow-y: auto;
    }
</style>