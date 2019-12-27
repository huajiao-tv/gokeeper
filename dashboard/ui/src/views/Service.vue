<template>
    <div>
        <v-toolbar flat>
            <v-toolbar-title>instances of - {{ currentService }}</v-toolbar-title>
            <v-divider
                    class="mx-2"
                    inset
                    vertical
            />
            <v-spacer/>
            <v-text-field
                    v-model="search"
                    append-icon="mdi-magnify"
                    label="Search"
                    single-line
                    hide-details
            />
            <v-menu offset-y>
                <template v-slot:activator="{ on }">
                    <v-btn v-on="on" color="primary" text>Service</v-btn>
                </template>
                <v-list>
                    <v-subheader>SELECT DOMAIN</v-subheader>
                    <v-list-item
                            v-for="item in services"
                            :key="item"
                            @click="getService(item)"
                    >
                        <!-- <v-list-item-icon>
                          <v-icon>mdi-buffer</v-icon>
                        </v-list-item-icon> -->
                        <v-list-item-content>
                            <v-list-item-title v-text="item" />
                        </v-list-item-content>
                    </v-list-item>
                </v-list>
            </v-menu>
            <!--<v-dialog v-model="dialog.item" persistent max-width="500px">
                <template v-slot:activator="{ on }">
                    <v-btn v-on="on" color="primary" text>New Item</v-btn>
                </template>
                <v-form ref="form" v-model="valid" lazy-validation>
                    <v-card>
                        <v-card-title>
                            <span class="headline">{{ formTitle }} - {{ currentCluster }}</span>
                        </v-card-title>
                        <v-card-text>
                            <v-container grid-list-md>
                                <v-layout wrap>
                                    <v-flex xs12 sm6>
                                        <v-combobox v-model="editedItem.file" :items="files" :readonly="!editable" :rules="emptyRules" label="File name" required/>
                                    </v-flex>
                                    <v-flex xs12 sm6>
                                        <v-text-field v-model="editedItem.section" :readonly="!editable" :rules="emptyRules" label="Section" required/>
                                    </v-flex>
                                    <v-flex xs12 sm6>
                                        <v-text-field v-model="editedItem.key" :readonly="!editable" :rules="emptyRules" label="Key" required/>
                                    </v-flex>
                                    <v-flex xs12 sm6>
                                        <v-select v-model="editedItem.type" :items="keyTypes" :readonly="!editable" :rules="emptyRules" label="Type" required/>
                                    </v-flex>
                                    <v-flex xs12>
                                        <v-textarea v-model="editedItem.value" :readonly="!editable" label="Value"/>
                                    </v-flex>
                                    <v-flex v-if="editable" xs12>
                                        <v-text-field v-model="editedItem.note" :rules="emptyRules" label="Comment" required/>
                                    </v-flex>
                                </v-layout>
                            </v-container>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer/>
                            <v-btn color="red darken-1" text @click.native="close">Cancel</v-btn>
                            <v-btn v-if="editable" color="blue darken-1" text @click.native="save">Save</v-btn>
                        </v-card-actions>
                    </v-card>
                </v-form>
            </v-dialog>-->



           <!-- <v-dialog v-model="dialog.domain" persistent max-width="500px">
                <template v-slot:activator="{ on }">
                    <v-btn v-on="on" color="primary" text>New Domain</v-btn>
                </template>
                <v-card>
                    <v-form ref="domainForm" v-model="valid" lazy-validation>
                        <v-card-title>
                            <span class="headline">New Domain</span>
                        </v-card-title>
                        <v-card-text>
                            <v-container grid-list-md>
                                <v-layout wrap>
                                    <v-flex xs12 sm6>
                                        <v-text-field
                                                v-model="newDomain.name"
                                                :rules="emptyRules"
                                                label="Domain Name"
                                                required
                                        />
                                    </v-flex>
                                    <v-flex xs12>
                                        <v-file-input
                                                v-model="newDomain.files"
                                                :rules="emptyRules"
                                                accept=".conf"
                                                show-size
                                                counter
                                                multiple
                                                label="Config Files (*.conf)"
                                        />
                                    </v-flex>
                                </v-layout>
                            </v-container>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn color="red darken-1" text @click="close()">Cancel</v-btn>
                            <v-btn color="blue darken-1" text @click="saveDomain">Save</v-btn>
                        </v-card-actions>
                    </v-form>
                </v-card>
            </v-dialog>-->



        </v-toolbar>

        <v-container fluid>
            <v-data-table
                    :headers="headers"
                    :items="instances"
                    :search="search"
                    :loading="loading"
                    :items-per-page="-1"
                    class="elevation-1"
            >
                <template v-slot:body="{ items }">
                    <tbody>
                    <tr v-for="item in items" :key="item.id">
                        <td>{{ item.id }}</td>
                        <td>{{ item.zone }}</td>
                        <td>{{ item.env }}</td>
                        <td>{{ item.hostname }}</td>
                        <td>{{ item.weight }}</td>
                        <td>
                            <table>
                                <tr v-for="(a,schema) in item.addr" :key="schema">
                                    <td>
                                        {{ schema }}
                                    </td>
                                    <td>
                                        {{ a }}
                                    </td>
                                </tr>
                            </table>
                        </td>

                    </tr>
                    </tbody>
                </template>
            </v-data-table>
        </v-container>
    </div>
</template>

<script>
    import axios from 'axios'
    import {getErrorMessage} from "@/tools/req"
    import {errorNotify} from "../tools/req";

    export default {
         data: () => ({
         dialog: {
             item:false,
             domain:false
         },
         search: '',
         loading: true,
         services: [],
         currentService: '',
         headers: [
             {text: 'ID', value: 'id', sortable: false},
             {text: 'Zone', value: 'zone', sortable: false},
             {text: 'Env', value: 'env', sortable: false},
             {text: 'Hostname', value: 'hostname', sortable: false},
             {text: 'Weight', value: 'weight', sortable: false},
             {text: 'Addr', value: 'addr', sortable: false},
         ],
         instances: [],
         editable: true,
         editedIndex: - 1
        ,
        editedItem: {
        }
        ,
        valid: true,
    }),

        computed: {
            formTitle() {
                return this.editedIndex === -1 ? 'New Item' : 'Edit Item'
            }
        },



        created() {
            this.initialize()
        },

        methods: {
            initialize() {
                axios.get('/discovery/services').then(res => {
                    this.services = res.data
                    if (this.services.length > 0) {
                        var service = this.services[0]
                        if (localStorage.service) {
                            service = localStorage.service
                        }
                        this.getService(service)
                    }
                }).catch(res=> {
                    errorNotify(getErrorMessage(res))
                    }
                )
            },


            getService(service) {
                this.loading = true
                this.desserts = []
                axios.get('/discovery/get/' + service).then(res => {
                    this.instances = []
                    for (var zone in res.data.instances){
                        let instances = res.data.instances[zone]
                        for (var id in instances) {
                            let instance = instances[id]
                            this.instances.push({
                                id: instance.id,
                                zone: (instance.zone===""||instance.zone==="unknown")?"default":instance.zone,
                                hostname: instance.hostname,
                                weight: instance.metadata["backend-metadata-zone_weight"] === undefined ? "-" : instance.metadata["backend-metadata-zone_weight"],
                                env: instance.env === "" ? "-" : instance.env,
                                addr: instance.addrs
                            })
                        }
                    }
                    this.currentService = service
                    localStorage.service = service
                    this.loading = false
                }).catch(res=>{
                    errorNotify(getErrorMessage(res))
                })
            },

        }
    }
</script>
