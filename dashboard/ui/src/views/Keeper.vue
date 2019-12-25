<template>
  <div>
    <v-toolbar flat>
      <v-toolbar-title>Keeper</v-toolbar-title>
      <v-divider class="mx-2" inset vertical />
      <v-spacer />
      <v-text-field v-model="search" append-icon="mdi-magnify" label="Search" single-line hide-details />
      <v-dialog v-model="dialog.keeper" persistent max-width="500px">
        <template v-slot:activator="{ on }">
          <v-btn v-on="on" color="primary" text>Add Keeper</v-btn>
        </template>
        <v-card>
          <v-form ref="keeperForm" v-model="valid" lazy-validation>
            <v-card-title>
              <span class="headline">Import Keeper</span>
            </v-card-title>
            <v-card-text>
              <v-container grid-list-md>
                <v-layout wrap>
                  <v-flex xs12>
                    <v-text-field
                      v-model="newKeeper"
                      :rules="emptyRules"
                      label="Keeper Address"
                      required
                    />
                  </v-flex>
                </v-layout>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer />
              <v-btn color="red darken-1" text @click="close('keeper')">Cancel</v-btn>
              <v-btn color="blue darken-1" text @click="saveKeeper">Save</v-btn>
            </v-card-actions>
          </v-form>
        </v-card>
      </v-dialog>
      <v-dialog v-model="dialog.domain" persistent max-width="500px">
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
                    <v-select
                      v-model="newDomain.keeper"
                      :items="keeperAddrs"
                      :rules="emptyRules"
                      label="Keeper"
                      required
                    />
                  </v-flex>
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
              <v-btn color="red darken-1" text @click="close('domain')">Cancel</v-btn>
              <v-btn color="blue darken-1" text @click="saveDomain">Save</v-btn>
            </v-card-actions>
          </v-form>
        </v-card>
      </v-dialog>
    </v-toolbar>

    <v-container fluid>
      <v-data-table
        :headers="headers"
        :items="desserts"
        :search="search"
        :loading="loading"
        :items-per-page="-1"
        class="elevation-1"
      />
    </v-container>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  data: () => ({
    dialog: {
      keeper: false,
      domain: false
    },
    search: '',
    loading: true,
    headers: [
      { text: 'Keeper', value: 'keeper', sortable: true },
      { text: 'Domain', value: 'domain', sortable: true }
    ],
    desserts: [],
    keeperAddrs: [],
    newDomain: {},
    newKeeper: '',
    valid: true,
    emptyRules: [v => !!v || 'Field is required']
  }),

  watch: {
    dialog(val) {
      val.keeper || this.close('keeper')
      val.domain || this.close('domain')
    }
  },

  created() {
    this.initialize()
  },

  methods: {
    initialize() {
      axios.get('/keeper/domains').then(res => {
        res.data.data.forEach(item => {
          this.desserts.push({
            keeper: item.keeper,
            domain: item.domain
          })
        })
        this.loading = false
      })
      axios.get('/keeper/index').then(res => {
        res.data.data.forEach(addr => {
          this.keeperAddrs.push(addr)
        })
      })
    },

    close(name) {
      if (name === 'keeper') {
        this.dialog.keeper = false
        this.newKeeper = ''
        this.$refs.keeperForm.reset()
      } else {
        this.dialog.domain = false
        this.$refs.domainForm.reset()
      }
    },

    saveKeeper() {
      if (!this.$refs.keeperForm.validate()) {
        return
      }
      axios.put('/keeper/' + this.newKeeper).then(res => {
        if (res.data.code !== 0) {
          console.log('add keeper failed', res.data)
          return
        }
        this.initialize()
        this.close('keeper')
      })
    },

    saveDomain() {
      if (!this.$refs.domainForm.validate()) {
        return
      }
      let formData = new FormData();
      for (let file of this.newDomain.files) {
        formData.append("files", file, file.name)
      }
      axios.put('/keeper/'+this.newDomain.keeper+'/'+this.newDomain.name, formData).then(res => {
        if (res.data.code !== 0) {
          console.log('add new domain failed', res.data)
          return
        }
        this.initialize()
        this.close('domain')
      })
    }
  }
}
</script>
